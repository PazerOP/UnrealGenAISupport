"""
Shared socket framing protocol for Unreal MCP communication.

Uses a 4-byte big-endian length prefix before each JSON payload.
This module MUST use only Python stdlib (no external deps)
because it runs both inside Unreal Engine's embedded Python
and in the standalone MCP server process.
"""
import struct
import json
import socket


HEADER_FORMAT = "!I"  # 4-byte unsigned int, big-endian (network byte order)
HEADER_SIZE = struct.calcsize(HEADER_FORMAT)  # 4 bytes
MAX_MESSAGE_SIZE = 16 * 1024 * 1024  # 16 MiB safety limit


def send_framed(sock: socket.socket, data: dict) -> None:
    """Serialize a dict as JSON and send it with a 4-byte length prefix."""
    payload = json.dumps(data).encode("utf-8")
    header = struct.pack(HEADER_FORMAT, len(payload))
    sock.sendall(header + payload)


def recv_framed(sock: socket.socket) -> dict | None:
    """Receive a length-prefixed JSON message and return the parsed dict.

    Returns None if the connection was closed cleanly before any data arrived.
    Raises ConnectionError or ValueError on protocol violations.
    """
    header = _recv_exact(sock, HEADER_SIZE)
    if header is None:
        return None  # Clean disconnect

    (length,) = struct.unpack(HEADER_FORMAT, header)
    if length > MAX_MESSAGE_SIZE:
        raise ValueError(
            f"Message size {length} exceeds maximum {MAX_MESSAGE_SIZE}"
        )

    payload = _recv_exact(sock, length)
    if payload is None:
        raise ConnectionError("Connection closed mid-message")

    return json.loads(payload.decode("utf-8"))


def _recv_exact(sock: socket.socket, num_bytes: int) -> bytes | None:
    """Receive exactly num_bytes from the socket.

    Returns None if the peer closed the connection before any bytes arrived.
    Raises ConnectionError if the connection drops after partial data.
    """
    data = bytearray()
    while len(data) < num_bytes:
        chunk = sock.recv(num_bytes - len(data))
        if not chunk:
            if not data:
                return None  # Clean close before any data
            raise ConnectionError(
                f"Connection closed after {len(data)}/{num_bytes} bytes"
            )
        data.extend(chunk)
    return bytes(data)
