# Content/Python/init_unreal.py - Using UE settings integration
import unreal
import importlib.util
import os
import sys
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

def start_mcp_server():
    """Start the external MCP server process (Go binary)"""
    global mcp_server_process
    try:
        # Find our plugin's Python directory to locate the plugin root
        plugin_python_path = None
        for path in sys.path:
            if "GenerativeAISupport/Content/Python" in path:
                plugin_python_path = path
                break

        if not plugin_python_path:
            log.log_error("Could not find plugin Python path")
            return False

        # Navigate up to the plugin root: Content/Python -> Content -> PluginRoot
        plugin_root = os.path.dirname(os.path.dirname(plugin_python_path))

        # Determine platform-specific binary name
        if sys.platform == 'win32':
            binary_name = "unreal-mcp-server.exe"
        elif sys.platform == 'darwin':
            binary_name = "unreal-mcp-server"
        else:
            binary_name = "unreal-mcp-server"

        mcp_binary_path = os.path.join(plugin_root, "McpServer", "bin", binary_name)

        if not os.path.exists(mcp_binary_path):
            log.log_error(f"MCP server binary not found at: {mcp_binary_path}")
            return False

        log.log_info(f"Starting MCP server binary: {mcp_binary_path}")

        # Pass the plugin root path as an environment variable so the Go binary
        # can locate knowledge_base/how_to_use.md
        env = os.environ.copy()
        env["UNREAL_PLUGIN_PATH"] = plugin_root

        # Create a detached process that will continue running
        # even if Unreal crashes (we'll handle proper shutdown with atexit)
        creationflags = 0
        if sys.platform == 'win32':
            creationflags = subprocess.CREATE_NEW_PROCESS_GROUP

        mcp_server_process = subprocess.Popen(
            [mcp_binary_path],
            env=env,
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
        # First get the class reference
        settings_class = unreal.load_class(None, '/Script/GenerativeAISupportEditor.GenerativeAISupportSettings')
        if settings_class:
            # Get the settings object using the class reference
            settings = unreal.get_default_object(settings_class)
            
            # Log available properties for debugging
            log.log_info(f"Settings object properties: {dir(settings)}")
            
            # Check if auto-start is enabled
            if hasattr(settings, 'auto_start_socket_server'):
                auto_start = settings.auto_start_socket_server
                log.log_info(f"Socket server auto-start setting: {auto_start}")
            else:
                log.log_warning("auto_start_socket_server property not found in settings")
                # Try alternative property names that might exist
                for prop in dir(settings):
                    if 'auto' in prop.lower() or 'socket' in prop.lower() or 'server' in prop.lower():
                        log.log_info(f"Found similar property: {prop}")
        else:
            log.log_error("Could not find GenerativeAISupportSettings class")
    except Exception as e:
        log.log_error(f"Error reading UE settings: {e}")
        log.log_info("Falling back to disabled auto-start")

    # Auto-start if configured
    if auto_start:
        log.log_info("Auto-starting Unreal Socket Server...")

        # Start Unreal Socket Server
        try:
            # Find our plugin's Python directory
            plugin_python_path = None
            for path in sys.path:
                if "GenerativeAISupport/Content/Python" in path:
                    plugin_python_path = path
                    break
    
            if plugin_python_path:
                server_path = os.path.join(plugin_python_path, "unreal_socket_server.py")
    
                if os.path.exists(server_path):
                    # Import and execute the server module
                    spec = importlib.util.spec_from_file_location("unreal_socket_server", server_path)
                    server_module = importlib.util.module_from_spec(spec)
                    spec.loader.exec_module(server_module)
                    log.log_info("Unreal Socket Server started successfully")

                    # Now start the MCP server
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
