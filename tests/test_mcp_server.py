"""Tests for the MCP server components that don't require Unreal Engine."""
import json
import os
import socket
import struct
import sys
import tempfile
import threading
import time
import unittest
from pathlib import Path
from unittest.mock import patch, MagicMock

# Ensure the Content/Python directory is on the path
TEST_DIR = Path(__file__).parent
PYTHON_DIR = TEST_DIR.parent / "Content" / "Python"
sys.path.insert(0, str(PYTHON_DIR))

from socket_protocol import (
    send_framed,
    recv_framed,
    _recv_exact,
    HEADER_FORMAT,
    HEADER_SIZE,
    MAX_MESSAGE_SIZE,
)


# ---------------------------------------------------------------------------
# socket_protocol tests
# ---------------------------------------------------------------------------
class TestSocketProtocol(unittest.TestCase):
    """Tests for the length-prefixed framing protocol."""

    def _make_pair(self):
        a, b = socket.socketpair()
        self.addCleanup(a.close)
        self.addCleanup(b.close)
        return a, b

    def test_round_trip_simple(self):
        a, b = self._make_pair()
        data = {"success": True, "message": "hello"}
        send_framed(a, data)
        result = recv_framed(b)
        self.assertEqual(result, data)

    def test_round_trip_nested(self):
        a, b = self._make_pair()
        data = {"actors": [{"name": "Cube", "loc": [1.0, 2.0, 3.0]}], "count": 1}
        send_framed(a, data)
        result = recv_framed(b)
        self.assertEqual(result, data)

    def test_round_trip_large_payload(self):
        a, b = self._make_pair()
        data = {"script": "x" * 200_000}
        send_framed(a, data)
        result = recv_framed(b)
        self.assertEqual(result, data)

    def test_round_trip_empty_dict(self):
        a, b = self._make_pair()
        send_framed(a, {})
        result = recv_framed(b)
        self.assertEqual(result, {})

    def test_round_trip_unicode(self):
        a, b = self._make_pair()
        data = {"text": "Hello, \u4e16\u754c! \U0001f600"}
        send_framed(a, data)
        result = recv_framed(b)
        self.assertEqual(result, data)

    def test_multiple_messages(self):
        a, b = self._make_pair()
        messages = [{"i": i} for i in range(10)]
        for msg in messages:
            send_framed(a, msg)
        for msg in messages:
            result = recv_framed(b)
            self.assertEqual(result, msg)

    def test_clean_disconnect_returns_none(self):
        a, b = self._make_pair()
        a.close()
        result = recv_framed(b)
        self.assertIsNone(result)

    def test_mid_message_disconnect_raises(self):
        a, b = self._make_pair()
        # Send just the header, then close
        payload = json.dumps({"x": 1}).encode("utf-8")
        header = struct.pack(HEADER_FORMAT, len(payload))
        a.sendall(header)
        a.close()
        with self.assertRaises(ConnectionError):
            recv_framed(b)

    def test_oversized_message_raises(self):
        a, b = self._make_pair()
        # Send a header claiming a message larger than MAX_MESSAGE_SIZE
        header = struct.pack(HEADER_FORMAT, MAX_MESSAGE_SIZE + 1)
        a.sendall(header)
        with self.assertRaises(ValueError) as ctx:
            recv_framed(b)
        self.assertIn("exceeds maximum", str(ctx.exception))

    def test_recv_exact_partial_reads(self):
        """Verify _recv_exact handles partial reads correctly."""
        a, b = self._make_pair()
        data = b"hello world"

        # Send in tiny chunks from a thread to force partial reads
        def send_slow():
            for byte in data:
                a.sendall(bytes([byte]))
                time.sleep(0.01)

        t = threading.Thread(target=send_slow)
        t.start()
        result = _recv_exact(b, len(data))
        t.join()
        self.assertEqual(result, data)

    def test_header_size_is_4(self):
        self.assertEqual(HEADER_SIZE, 4)


# ---------------------------------------------------------------------------
# socket_protocol integration: client/server simulation
# ---------------------------------------------------------------------------
class TestClientServerSimulation(unittest.TestCase):
    """Simulate the MCP server <-> Unreal socket server communication."""

    def test_command_dispatch_round_trip(self):
        """Simulate: MCP sends command, server responds."""
        server_sock = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        server_sock.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        server_sock.bind(("localhost", 0))  # OS-assigned port
        port = server_sock.getsockname()[1]
        server_sock.listen(1)

        response_from_server = {"success": True, "actor_name": "Cube_1"}

        def server_thread():
            conn, _ = server_sock.accept()
            try:
                command = recv_framed(conn)
                # "Process" the command and send response
                self.assertEqual(command["type"], "spawn")
                send_framed(conn, response_from_server)
            finally:
                conn.close()
                server_sock.close()

        t = threading.Thread(target=server_thread)
        t.start()

        # Client side (simulates send_to_unreal)
        client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        client.connect(("localhost", port))
        try:
            send_framed(client, {"type": "spawn", "actor_class": "Cube"})
            result = recv_framed(client)
        finally:
            client.close()

        t.join(timeout=5)
        self.assertEqual(result, response_from_server)


# ---------------------------------------------------------------------------
# PID file tests
# ---------------------------------------------------------------------------
class TestPidFile(unittest.TestCase):
    """Test PID file creation from mcp_server.py."""

    def test_write_pid_file_creates_and_cleans_up(self):
        # We need to mock the mcp import since it's not installed in CI
        with self._mock_mcp_imports():
            import mcp_server

            with patch.dict(os.environ, {"UNREAL_PORT": "12345"}):
                pid_path = mcp_server.write_pid_file()

            self.assertIsNotNone(pid_path)
            self.assertTrue(os.path.exists(pid_path))

            with open(pid_path) as f:
                content = f.read()
            lines = content.strip().split("\n")
            self.assertEqual(int(lines[0]), os.getpid())
            self.assertEqual(lines[1], "12345")

            # Clean up
            if os.path.exists(pid_path):
                os.remove(pid_path)

    @staticmethod
    def _mock_mcp_imports():
        """Context manager that mocks the mcp package imports."""
        mock_fastmcp = MagicMock()
        mock_fastmcp.FastMCP = MagicMock(return_value=MagicMock())
        mock_fastmcp.Image = MagicMock()
        return patch.dict(sys.modules, {
            "mcp": MagicMock(),
            "mcp.server": MagicMock(),
            "mcp.server.fastmcp": mock_fastmcp,
        })


# ---------------------------------------------------------------------------
# MCP tool function tests (with mocked send_to_unreal)
# ---------------------------------------------------------------------------
class TestMCPToolFunctions(unittest.TestCase):
    """Test MCP tool wrapper functions with mocked socket communication."""

    @classmethod
    def setUpClass(cls):
        """Import mcp_server with mocked mcp dependency."""
        mock_fastmcp = MagicMock()
        # Make @mcp.tool() a no-op decorator
        mock_mcp_instance = MagicMock()
        mock_mcp_instance.tool.return_value = lambda fn: fn
        mock_fastmcp.FastMCP.return_value = mock_mcp_instance
        mock_fastmcp.Image = MagicMock()

        patcher = patch.dict(sys.modules, {
            "mcp": MagicMock(),
            "mcp.server": MagicMock(),
            "mcp.server.fastmcp": mock_fastmcp,
        })
        patcher.start()
        cls._patcher = patcher

        # Force reimport to pick up mocks
        if "mcp_server" in sys.modules:
            del sys.modules["mcp_server"]
        import mcp_server
        cls.mod = mcp_server

    @classmethod
    def tearDownClass(cls):
        cls._patcher.stop()
        if "mcp_server" in sys.modules:
            del sys.modules["mcp_server"]

    def test_spawn_object_success(self):
        with patch.object(self.mod, "send_to_unreal", return_value={"success": True}):
            result = self.mod.spawn_object("Cube")
            self.assertIn("Successfully spawned Cube", result)

    def test_spawn_object_with_label(self):
        with patch.object(self.mod, "send_to_unreal", return_value={"success": True}):
            result = self.mod.spawn_object("Sphere", actor_label="MySphere")
            self.assertIn("MySphere", result)

    def test_spawn_object_failure(self):
        with patch.object(self.mod, "send_to_unreal", return_value={"success": False, "error": "class not found"}):
            result = self.mod.spawn_object("BadClass")
            self.assertIn("Failed", result)
            self.assertIn("Hint", result)

    def test_spawn_object_default_args_are_safe(self):
        """Verify mutable default args don't leak between calls."""
        with patch.object(self.mod, "send_to_unreal", return_value={"success": True}) as mock:
            self.mod.spawn_object("Cube")
            call1_args = mock.call_args[0][0]
            self.mod.spawn_object("Sphere")
            call2_args = mock.call_args[0][0]
            # Each call should get fresh defaults
            self.assertEqual(call1_args["location"], [0, 0, 0])
            self.assertEqual(call2_args["location"], [0, 0, 0])

    def test_execute_python_script_success(self):
        with patch.object(self.mod, "send_to_unreal", return_value={"success": True, "output": "hello"}):
            result = self.mod.execute_python_script("print('hello')")
            self.assertIn("hello", result)

    def test_execute_python_script_failure_includes_partial_output(self):
        with patch.object(self.mod, "send_to_unreal", return_value={
            "success": False, "error": "NameError", "output": "partial"
        }):
            result = self.mod.execute_python_script("bad_code")
            self.assertIn("NameError", result)
            self.assertIn("partial", result)

    def test_execute_unreal_command_blocks_py_prefix(self):
        result = self.mod.execute_unreal_command("py some_script.py")
        self.assertIn("execute_python_script", result)

    def test_create_game_mode_uses_response_directly(self):
        """Verify create_game_mode doesn't json.loads a dict (the old bug)."""
        with patch.object(self.mod, "send_to_unreal", return_value={
            "success": True, "message": "Game mode created"
        }):
            result = self.mod.create_game_mode("/Game/GM", "/Game/BP_Pawn")
            self.assertIn("Game mode created", result)

    def test_add_component_with_events_uses_response_directly(self):
        """Verify add_component_with_events doesn't json.loads a dict."""
        with patch.object(self.mod, "send_to_unreal", return_value={
            "success": True, "message": "Added component",
            "events": {"begin_guid": "ABC", "end_guid": "DEF"}
        }):
            result = self.mod.add_component_with_events("/Game/BP", "Box", "BoxComponent")
            self.assertIn("Added component", result)
            self.assertIn("ABC", result)

    def test_edit_widget_property_uses_response_directly(self):
        """Verify edit_widget_property doesn't json.loads a dict."""
        with patch.object(self.mod, "send_to_unreal", return_value={
            "success": True, "message": "Property set"
        }):
            result = self.mod.edit_widget_property("/Game/UI/WBP", "Text1", "Text", "Hello")
            self.assertIn("Property set", result)

    def test_add_widget_uses_response_directly(self):
        """Verify add_widget_to_user_widget doesn't json.loads a dict."""
        with patch.object(self.mod, "send_to_unreal", return_value={
            "success": True, "message": "Widget added", "widget_name": "MyText"
        }):
            result = self.mod.add_widget_to_user_widget("/Game/UI/WBP", "TextBlock", "MyText")
            self.assertIn("Widget added", result)

    def test_connect_blueprint_nodes_has_docstring(self):
        self.assertIsNotNone(self.mod.connect_blueprint_nodes.__doc__)
        self.assertIn("source_pin", self.mod.connect_blueprint_nodes.__doc__)

    def test_send_to_unreal_reads_env_vars(self):
        """Verify send_to_unreal uses UNREAL_HOST and UNREAL_PORT env vars."""
        import inspect
        source = inspect.getsource(self.mod.send_to_unreal)
        self.assertIn("UNREAL_HOST", source)
        self.assertIn("UNREAL_PORT", source)

    def test_main_function_exists(self):
        self.assertTrue(callable(getattr(self.mod, "main", None)))


# ---------------------------------------------------------------------------
# Syntax validation
# ---------------------------------------------------------------------------
class TestSyntax(unittest.TestCase):
    """Verify all Python files have valid syntax."""

    def test_all_python_files_parse(self):
        import ast

        python_dir = PYTHON_DIR
        errors = []
        for py_file in python_dir.rglob("*.py"):
            try:
                with open(py_file) as f:
                    ast.parse(f.read())
            except SyntaxError as e:
                errors.append(f"{py_file.relative_to(python_dir)}: {e}")

        if errors:
            self.fail("Syntax errors found:\n" + "\n".join(errors))


# ---------------------------------------------------------------------------
# pyproject.toml validation
# ---------------------------------------------------------------------------
class TestPackaging(unittest.TestCase):
    """Verify the pyproject.toml is well-formed."""

    def test_pyproject_toml_valid(self):
        import tomllib

        toml_path = PYTHON_DIR / "pyproject.toml"
        self.assertTrue(toml_path.exists(), "pyproject.toml not found")

        with open(toml_path, "rb") as f:
            config = tomllib.load(f)

        self.assertIn("project", config)
        self.assertIn("name", config["project"])
        self.assertIn("dependencies", config["project"])
        self.assertIn("scripts", config["project"])

        # Verify mcp dependency is declared
        deps = config["project"]["dependencies"]
        self.assertTrue(any("mcp" in d for d in deps), f"mcp not in deps: {deps}")

        # Verify entry point
        self.assertIn("unreal-mcp-server", config["project"]["scripts"])

    def test_pep723_metadata_present(self):
        mcp_server_path = PYTHON_DIR / "mcp_server.py"
        content = mcp_server_path.read_text()
        self.assertIn("# /// script", content)
        self.assertIn("mcp[cli]", content)
        self.assertIn("# ///", content)


if __name__ == "__main__":
    unittest.main()
