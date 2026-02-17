# Content/Python/init_unreal.py - Using UE settings integration
import unreal
import importlib.util
import os
import sys
import shutil
import subprocess
import atexit
from utils import logging as log

# Global process handle for MCP server
mcp_server_process = None


def shutdown_mcp_server():
    """Shutdown the MCP server process when Unreal Editor closes"""
    global mcp_server_process
    if mcp_server_process:
        log.log_info("Shutting down MCP server process...")
        try:
            mcp_server_process.terminate()
            mcp_server_process = None
            log.log_info("MCP server process terminated successfully")
        except Exception as e:
            log.log_error(f"Error terminating MCP server: {e}")


def _find_plugin_python_path():
    """Find the plugin's Content/Python directory from sys.path."""
    for path in sys.path:
        if "GenerativeAISupport/Content/Python" in path:
            return path
    return None


def _find_mcp_command(mcp_server_path):
    """Determine the best command to launch the MCP server.

    Priority:
    1. uv (handles deps automatically via PEP 723 inline metadata)
    2. System python3/python (user must have mcp[cli] installed)
    3. sys.executable (Unreal's embedded Python — unlikely to have mcp)
    """
    # Try uv first — it reads inline script metadata and auto-installs deps
    uv_path = shutil.which("uv")
    if uv_path:
        log.log_info(f"Found uv at: {uv_path}")
        return [uv_path, "run", mcp_server_path]

    # Try system Python (not Unreal's embedded one)
    for name in ("python3", "python"):
        python_path = shutil.which(name)
        if python_path and python_path != sys.executable:
            log.log_info(f"Found system Python at: {python_path}")
            return [python_path, mcp_server_path]

    # Last resort: Unreal's embedded Python (probably won't have mcp[cli])
    log.log_warning(
        f"Could not find uv or system Python. Falling back to Unreal's Python: {sys.executable}. "
        f"This may fail if mcp[cli] is not installed in Unreal's Python environment."
    )
    return [sys.executable, mcp_server_path]


def start_mcp_server():
    """Start the external MCP server process"""
    global mcp_server_process
    try:
        plugin_python_path = _find_plugin_python_path()
        if not plugin_python_path:
            log.log_error("Could not find plugin Python path")
            return False

        mcp_server_path = os.path.join(plugin_python_path, "mcp_server.py")
        if not os.path.exists(mcp_server_path):
            log.log_error(f"MCP server script not found at: {mcp_server_path}")
            return False

        cmd = _find_mcp_command(mcp_server_path)
        log.log_info(f"Starting MCP server: {' '.join(cmd)}")

        creationflags = 0
        if sys.platform == 'win32':
            creationflags = subprocess.CREATE_NEW_PROCESS_GROUP

        mcp_server_process = subprocess.Popen(
            cmd,
            creationflags=creationflags,
            stdout=subprocess.PIPE,
            stderr=subprocess.PIPE,
            text=True
        )

        log.log_info(f"MCP server started with PID: {mcp_server_process.pid}")

        # Register cleanup handler to ensure process is terminated when Unreal exits
        atexit.register(shutdown_mcp_server)

        return True
    except Exception as e:
        log.log_error(f"Error starting MCP server: {e}")
        return False


def initialize_socket_server():
    """
    Initialize the socket server if auto-start is enabled in UE settings
    """
    auto_start = False

    # Get settings from UE settings system
    try:
        settings_class = unreal.load_class(None, '/Script/GenerativeAISupportEditor.GenerativeAISupportSettings')
        if settings_class:
            settings = unreal.get_default_object(settings_class)

            if hasattr(settings, 'auto_start_socket_server'):
                auto_start = settings.auto_start_socket_server
                log.log_info(f"Socket server auto-start setting: {auto_start}")
            else:
                log.log_warning("auto_start_socket_server property not found in settings")
                for prop in dir(settings):
                    if 'auto' in prop.lower() or 'socket' in prop.lower() or 'server' in prop.lower():
                        log.log_info(f"Found similar property: {prop}")
        else:
            log.log_error("Could not find GenerativeAISupportSettings class")
    except Exception as e:
        log.log_error(f"Error reading UE settings: {e}")
        log.log_info("Falling back to disabled auto-start")

    if auto_start:
        log.log_info("Auto-starting Unreal Socket Server...")

        try:
            plugin_python_path = _find_plugin_python_path()
            if plugin_python_path:
                server_path = os.path.join(plugin_python_path, "unreal_socket_server.py")

                if os.path.exists(server_path):
                    spec = importlib.util.spec_from_file_location("unreal_socket_server", server_path)
                    server_module = importlib.util.module_from_spec(spec)
                    spec.loader.exec_module(server_module)
                    log.log_info("Unreal Socket Server started successfully")

                    if start_mcp_server():
                        log.log_info("Both servers started successfully")
                    else:
                        log.log_error("Failed to start MCP server")
                else:
                    log.log_error(f"Server file not found at: {server_path}")
            else:
                log.log_error("Could not find plugin Python path")
        except Exception as e:
            log.log_error(f"Error starting socket server: {e}")
    else:
        log.log_info("Unreal Socket Server auto-start is disabled")

# Run initialization when this script is loaded
initialize_socket_server()
