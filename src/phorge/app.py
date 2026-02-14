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
        Binding("ctrl+r", "refresh", "Refresh", show=True),
        Binding("ctrl+e", "edit_config", "Config", show=True),
    ]

    COMMANDS = App.COMMANDS | {ServerCommandProvider}

    forge_client: ForgeClient | None = None

    def on_mount(self) -> None:
        config = load_config()
        if not config.forge.api_key:
            from phorge.screens.setup import SetupScreen
            self.push_screen(SetupScreen())
        else:
            self.forge_client = ForgeClient(config.forge.api_key)
            from phorge.screens.main import MainScreen
            self.push_screen(MainScreen())

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

    def action_ssh_selected(self) -> None:
        """SSH to the currently selected server in the tree."""
        import subprocess
        from phorge.screens.main import MainScreen
        from phorge.widgets.server_tree import ServerTree

        screen = self.screen
        if not isinstance(screen, MainScreen):
            return

        tree = screen.query_one(ServerTree)
        node = tree.cursor_node
        if node is None or node.data is None:
            self.notify("No server selected", severity="warning")
            return

        # Walk up to find the server IP
        ip = node.data.server_ip
        port = node.data.ssh_port
        if not ip:
            # Try walking up parents
            current = node
            while current.parent is not None:
                current = current.parent
                if current.data and current.data.server_ip:
                    ip = current.data.server_ip
                    port = current.data.ssh_port
                    break

        if ip:
            with self.suspend():
                subprocess.run(["ssh", "-p", str(port), f"forge@{ip}"])
        else:
            self.notify("Could not determine server IP", severity="error")
