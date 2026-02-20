"""Main Phorge application class."""

from __future__ import annotations

from pathlib import Path

from textual import work
from textual.app import App, ComposeResult
from textual.binding import Binding

from phorge.api.client import ForgeClient
from phorge.commands.palette import ServerCommandProvider
from phorge.config import load_config


class PhorgeApp(App):
    """Phorge - Laravel Forge TUI."""

    TITLE = "Phorge"
    SUB_TITLE = "Laravel Forge TUI"

    CSS_PATH = [
        Path(__file__).parent / "styles" / "app.tcss",
        Path(__file__).parent / "styles" / "panels.tcss",
        Path(__file__).parent / "styles" / "modals.tcss",
    ]

    BINDINGS = [
        Binding("ctrl+q", "quit", "Quit", show=True, priority=True),
        Binding("ctrl+p", "command_palette", "Commands", show=True),
        Binding("ctrl+s", "ssh_selected", "SSH", show=True),
        Binding("ctrl+f", "sftp_selected", "Files", show=True),
        Binding("ctrl+d", "db_selected", "Database", show=True),
        Binding("ctrl+r", "refresh", "Refresh", show=True),
        Binding("ctrl+g", "switch_server", "Servers", show=True),
        Binding("ctrl+e", "edit_config", "Config", show=True),
    ]

    COMMANDS = App.COMMANDS | {ServerCommandProvider}

    forge_client: ForgeClient | None = None
    ssh_user: str = "forge"
    server_users: dict[str, str] = {}

    def on_mount(self) -> None:
        config = load_config()
        if not config.forge.api_key:
            from phorge.screens.setup import SetupScreen
            self.push_screen(SetupScreen())
        else:
            self.forge_client = ForgeClient(config.forge.api_key)
            self.ssh_user = config.forge.ssh_user
            self.server_users = dict(config.server_users)
            from phorge.screens.main import MainScreen
            self.push_screen(MainScreen())

    def get_ssh_user(self, server_id: int) -> str:
        """Get the SSH user for a server, falling back to the global default."""
        return self.server_users.get(str(server_id), self.ssh_user)

    async def on_unmount(self) -> None:
        if self.forge_client:
            await self.forge_client.close()

    def action_edit_config(self) -> None:
        from phorge.screens.config_screen import ConfigScreen
        self.push_screen(ConfigScreen())

    def action_refresh(self) -> None:
        from phorge.screens.main import MainScreen
        screen = self.screen
        if isinstance(screen, MainScreen):
            screen.action_refresh()

    def action_switch_server(self) -> None:
        from phorge.screens.main import MainScreen
        screen = self.screen
        if isinstance(screen, MainScreen):
            screen.action_switch_server()

    def _resolve_server_info(self):
        """Resolve server ID, IP, port, SSH user, and site directory from the selected tree node."""
        from phorge.screens.main import MainScreen
        from phorge.widgets.server_tree import ServerTree

        screen = self.screen
        if not isinstance(screen, MainScreen):
            return None

        tree = screen.query_one(ServerTree)
        node = tree.cursor_node
        if node is None or node.data is None:
            self.notify("No server selected", severity="warning")
            return None

        server_id = node.data.server_id
        ip = node.data.server_ip
        port = node.data.ssh_port
        if not ip:
            current = node
            while current.parent is not None:
                current = current.parent
                if current.data and current.data.server_ip:
                    server_id = current.data.server_id
                    ip = current.data.server_ip
                    port = current.data.ssh_port
                    break

        if not ip:
            self.notify("Could not determine server IP", severity="error")
            return None

        user = self.get_ssh_user(server_id)

        site_dir = node.data.site_directory
        if not site_dir:
            current = node
            while current.parent is not None:
                current = current.parent
                if current.data and current.data.site_directory:
                    site_dir = current.data.site_directory
                    break

        return user, ip, port, site_dir

    def action_ssh_selected(self) -> None:
        """SSH to the currently selected server in the tree."""
        import subprocess

        info = self._resolve_server_info()
        if not info:
            return

        user, ip, port, site_dir = info
        cmd = ["ssh", "-t", "-p", str(port), f"{user}@{ip}"]
        if site_dir:
            cmd.append(f"cd {site_dir} && exec $SHELL -l")

        with self.suspend():
            subprocess.call(cmd)

    def action_sftp_selected(self) -> None:
        """Open termscp SFTP session to the currently selected server."""
        import shutil
        import subprocess

        if not shutil.which("termscp"):
            self.notify("termscp is not installed, please install it first.", severity="error")
            return

        info = self._resolve_server_info()
        if not info:
            return

        user, ip, port, site_dir = info
        address = f"sftp://{user}@{ip}:{port}:{site_dir or f'/home/{user}'}"

        with self.suspend():
            subprocess.call(["termscp", address])

    def action_db_selected(self) -> None:
        """Open a lazysql database session via SSH tunnel."""
        self._launch_db_connection()

    @work(exclusive=True, group="db-connect")
    async def _launch_db_connection(self) -> None:
        """Resolve DB credentials from .env, open SSH tunnel, and launch lazysql."""
        import shutil
        import socket
        import subprocess

        from phorge.api.endpoints.environment import EnvironmentAPI
        from phorge.screens.main import MainScreen
        from phorge.utils.env_parser import parse_env
        from phorge.widgets.server_tree import ServerTree

        if not shutil.which("lazysql"):
            self.notify(
                "lazysql is not installed, please install it first.",
                severity="error",
            )
            return

        info = self._resolve_server_info()
        if not info:
            return

        user, ip, ssh_port, site_dir = info

        # Find site_id by walking up the tree
        screen = self.screen
        if not isinstance(screen, MainScreen):
            return

        tree = screen.query_one(ServerTree)
        node = tree.cursor_node
        site_id = None
        server_id = None

        if node is not None and node.data is not None:
            site_id = node.data.site_id
            server_id = node.data.server_id
            if not site_id:
                current = node
                while current.parent is not None:
                    current = current.parent
                    if current.data and current.data.site_id:
                        site_id = current.data.site_id
                        server_id = current.data.server_id
                        break

        if not site_id:
            self.notify(
                "Select a site first to read database credentials from its .env file",
                severity="warning",
            )
            return

        self.notify("Connecting to database...", severity="information")

        # Fetch and parse .env
        try:
            api = EnvironmentAPI(self.forge_client)
            env_content = await api.get(server_id, site_id)
        except Exception as exc:
            self.notify(f"Failed to fetch .env: {exc}", severity="error")
            return

        env = parse_env(env_content)
        db_connection = env.get("DB_CONNECTION", "mysql")
        is_postgres = db_connection in ("pgsql", "postgres", "postgresql")
        db_host = env.get("DB_HOST", "127.0.0.1")
        db_name = env.get("DB_DATABASE", "")
        db_user = env.get("DB_USERNAME", "")
        db_pass = env.get("DB_PASSWORD", "")

        if is_postgres:
            db_port = env.get("DB_PORT", "5432")
        else:
            db_port = env.get("DB_PORT", "3306")

        if not db_name:
            self.notify("No DB_DATABASE found in .env", severity="error")
            return

        # Find a free local port for the SSH tunnel
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
            s.bind(("127.0.0.1", 0))
            local_port = s.getsockname()[1]

        # Build SSH tunnel command
        tunnel_cmd = [
            "ssh", "-N", "-L",
            f"{local_port}:{db_host}:{db_port}",
            "-p", str(ssh_port),
            f"{user}@{ip}",
        ]

        tunnel_proc = subprocess.Popen(tunnel_cmd, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)
        try:
            # Give the tunnel a moment to establish
            import asyncio
            await asyncio.sleep(1)

            if tunnel_proc.poll() is not None:
                self.notify(
                    "SSH tunnel failed to establish. Check your SSH configuration.",
                    severity="error",
                )
                return

            # Build lazysql connection URL
            from urllib.parse import quote

            if is_postgres:
                url = f"postgres://{quote(db_user, safe='')}:{quote(db_pass, safe='')}@localhost:{local_port}/{db_name}"
            else:
                url = f"mysql://{quote(db_user, safe='')}:{quote(db_pass, safe='')}@localhost:{local_port}/{db_name}"

            with self.suspend():
                subprocess.call(["lazysql", url])
        finally:
            tunnel_proc.terminate()
            tunnel_proc.wait()
