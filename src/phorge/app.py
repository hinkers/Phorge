"""Main Phorge application class."""

from __future__ import annotations

from pathlib import Path

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
