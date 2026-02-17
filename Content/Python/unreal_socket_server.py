import socket
import json
import queue
import unreal
import threading
import time
from typing import Dict, Any, Tuple, List, Optional

# Import handlers
from handlers import basic_commands, actor_commands, blueprint_commands, python_commands
from handlers import ui_commands
from utils import logging as log
from socket_protocol import send_framed, recv_framed

# Global queues and state
command_queue = queue.Queue()
response_dict = {}

# Shutdown coordination
_shutdown_flag = threading.Event()
_server_thread = None


class CommandDispatcher:
    """
    Dispatches commands to appropriate handlers based on command type
    """
    def __init__(self):
        # Register command handlers
        self.handlers = {
            "handshake": self._handle_handshake,

            # Basic object commands
            "spawn": basic_commands.handle_spawn,
            "create_material": basic_commands.handle_create_material,
            "modify_object": actor_commands.handle_modify_object,
            "create_game_mode": actor_commands.handle_create_game_mode,
            "take_screenshot": basic_commands.handle_take_screenshot,

            # Blueprint commands
            "create_blueprint": blueprint_commands.handle_create_blueprint,
            "add_component": blueprint_commands.handle_add_component,
            "add_variable": blueprint_commands.handle_add_variable,
            "add_function": blueprint_commands.handle_add_function,
            "add_node": blueprint_commands.handle_add_node,
            "connect_nodes": blueprint_commands.handle_connect_nodes,
            "compile_blueprint": blueprint_commands.handle_compile_blueprint,
            "spawn_blueprint": blueprint_commands.handle_spawn_blueprint,
            "delete_node": blueprint_commands.handle_delete_node,

            # Getters
            "get_node_guid": blueprint_commands.handle_get_node_guid,
            "get_all_nodes": blueprint_commands.handle_get_all_nodes,
            "get_node_suggestions": blueprint_commands.handle_get_node_suggestions,

            # Bulk commands
            "add_nodes_bulk": blueprint_commands.handle_add_nodes_bulk,
            "connect_nodes_bulk": blueprint_commands.handle_connect_nodes_bulk,

            # Python and console
            "execute_python": python_commands.handle_execute_python,
            "execute_unreal_command": python_commands.handle_execute_unreal_command,

            # Actor/component editing
            "edit_component_property": actor_commands.handle_edit_component_property,
            "add_component_with_events": actor_commands.handle_add_component_with_events,

            # Scene
            "get_all_scene_objects": basic_commands.handle_get_all_scene_objects,
            "create_project_folder": basic_commands.handle_create_project_folder,
            "get_files_in_folder": basic_commands.handle_get_files_in_folder,

            # Input
            "add_input_binding": basic_commands.handle_add_input_binding,

            # UI commands
            "add_widget_to_user_widget": ui_commands.handle_add_widget_to_user_widget,
            "edit_widget_property": ui_commands.handle_edit_widget_property,
        }

    def dispatch(self, command: Dict[str, Any]) -> Dict[str, Any]:
        """Dispatch command to appropriate handler"""
        command_type = command.get("type")
        if command_type not in self.handlers:
            return {"success": False, "error": f"Unknown command type: {command_type}"}

        try:
            handler = self.handlers[command_type]
            return handler(command)
        except Exception as e:
            log.log_error(f"Error processing command: {str(e)}")
            return {"success": False, "error": str(e)}

    def _handle_handshake(self, command: Dict[str, Any]) -> Dict[str, Any]:
        """Built-in handler for handshake command"""
        message = command.get("message", "")
        log.log_info(f"Handshake received: {message}")

        # Get Unreal Engine version
        engine_version = unreal.SystemLibrary.get_engine_version()

        # Add connection and session information
        connection_info = {
            "status": "Connected",
            "engine_version": engine_version,
            "timestamp": time.time(),
            "session_id": f"UE-{int(time.time())}"
        }

        return {
            "success": True,
            "message": f"Received: {message}",
            "connection_info": connection_info
        }


# Create global dispatcher instance
dispatcher = CommandDispatcher()


def process_commands(delta_time=None):
    """Process commands on the main thread (called via Unreal tick)"""
    try:
        command_id, command = command_queue.get_nowait()
    except queue.Empty:
        return

    log.log_info(f"Processing command on main thread: {command}")

    try:
        response = dispatcher.dispatch(command)
        response_dict[command_id] = response
    except Exception as e:
        log.log_error(f"Error processing command: {str(e)}", include_traceback=True)
        response_dict[command_id] = {"success": False, "error": str(e)}


def socket_server_thread(port=9877):
    """Socket server running in a separate thread"""
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server_socket.bind(('localhost', port))
    server_socket.listen(1)
    server_socket.settimeout(1.0)  # Allow periodic check of shutdown flag
    log.log_info(f"Unreal Engine socket server started on port {port}")

    command_counter = 0

    while not _shutdown_flag.is_set():
        conn = None
        try:
            try:
                conn, addr = server_socket.accept()
            except socket.timeout:
                continue  # Check shutdown flag and loop

            conn.settimeout(5)

            command = recv_framed(conn)
            if command is None:
                log.log_warning("Client disconnected before sending data")
                continue

            log.log_info(f"Received command: {command}")

            # For handshake, respond directly from the thread
            if command.get("type") == "handshake":
                response = dispatcher.dispatch(command)
                send_framed(conn, response)
            else:
                # Queue for main thread execution
                command_id = command_counter
                command_counter += 1
                command_queue.put((command_id, command))

                # Wait for the response with a timeout
                timeout = 30  # seconds
                start_time = time.time()
                while command_id not in response_dict and time.time() - start_time < timeout:
                    time.sleep(0.1)

                if command_id in response_dict:
                    response = response_dict.pop(command_id)
                    send_framed(conn, response)
                else:
                    send_framed(conn, {"success": False, "error": "Command timed out"})

        except Exception as e:
            log.log_error(f"Error in socket server: {str(e)}", include_traceback=True)
            if conn is not None:
                try:
                    send_framed(conn, {"success": False, "error": str(e)})
                except Exception:
                    pass
        finally:
            if conn is not None:
                try:
                    conn.close()
                except Exception:
                    pass

    server_socket.close()
    log.log_info("Socket server shut down")


# Register tick function to process commands on main thread
def register_command_processor():
    """Register the command processor with Unreal's tick system"""
    unreal.register_slate_post_tick_callback(process_commands)
    log.log_info("Command processor registered")


def shutdown_server():
    """Signal the socket server to shut down gracefully."""
    _shutdown_flag.set()
    if _server_thread is not None:
        _server_thread.join(timeout=5.0)
    log.log_info("Socket server shutdown complete")


# Initialize the server
def initialize_server(port=9877):
    """Initialize and start the socket server"""
    global _server_thread
    _shutdown_flag.clear()

    _server_thread = threading.Thread(target=socket_server_thread, args=(port,))
    _server_thread.daemon = True
    _server_thread.start()
    log.log_info("Socket server thread started")

    # Register the command processor on the main thread
    register_command_processor()

    log.log_info("Unreal Engine AI command server initialized successfully")

# Auto-start the server when this module is imported
initialize_server()
