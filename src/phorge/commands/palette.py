"""Command palette providers for quick navigation."""

from __future__ import annotations

import subprocess
import webbrowser

from textual.command import Provider, Hit, Hits, DiscoveryHit

from phorge.api.endpoints.servers import ServersAPI
from phorge.api.models import Server


class ServerCommandProvider(Provider):
    """Provides server/site navigation and action commands."""

    async def startup(self) -> None:
        """Cache server list when palette opens."""
        self._servers: list[Server] = []
        client = getattr(self.app, "forge_client", None)
        if client is None:
            return
        try:
            api = ServersAPI(client)
            self._servers = await api.list()
        except Exception:
            pass

    async def discover(self) -> Hits:
        """Show available commands when palette first opens."""
        # Config command
        yield DiscoveryHit(
            "Edit Configuration",
            self._open_config,
            help="Edit API key, editor, and UI settings",
        )

        # Server commands
        for server in self._servers:
            ip = server.ip_address or "no ip"
            yield DiscoveryHit(
                f"Server: {server.name} ({ip})",
                self._navigate_to_server(server),
                help=f"Region: {server.region or 'N/A'} | Provider: {server.provider or 'N/A'}",
            )
            yield DiscoveryHit(
                f"SSH to {server.name}",
                self._ssh_to_server(server),
                help=f"Connect via SSH to {ip}",
            )

    async def search(self, query: str) -> Hits:
        """Fuzzy search across servers and actions."""
        matcher = self.matcher(query)

        # Config
        config_label = "Edit Configuration"
        score = matcher.match(config_label)
        if score > 0:
            yield Hit(score, matcher.highlight(config_label), self._open_config)

        # Refresh
        refresh_label = "Refresh Servers"
        score = matcher.match(refresh_label)
        if score > 0:
            yield Hit(score, matcher.highlight(refresh_label), self._refresh)

        for server in self._servers:
            ip = server.ip_address or ""

            # Navigate to server
            label = f"Server: {server.name} ({ip})"
            score = matcher.match(label)
            if score > 0:
                yield Hit(
                    score,
                    matcher.highlight(label),
                    self._navigate_to_server(server),
                    help=f"Region: {server.region or 'N/A'}",
                )

            # SSH to server
            ssh_label = f"SSH to {server.name}"
            score = matcher.match(ssh_label)
            if score > 0:
                yield Hit(
                    score,
                    matcher.highlight(ssh_label),
                    self._ssh_to_server(server),
                    help=f"Connect to {ip}",
                )

    def _open_config(self) -> None:
        from phorge.screens.config_screen import ConfigScreen
        self.app.push_screen(ConfigScreen())

    def _refresh(self) -> None:
        from phorge.screens.main import MainScreen
        screen = self.app.screen
        if isinstance(screen, MainScreen):
            screen.action_refresh()

    def _navigate_to_server(self, server: Server):
        def callback() -> None:
            from phorge.screens.main import MainScreen
            from phorge.widgets.server_tree import ServerTree, NodeType

            screen = self.app.screen
            if not isinstance(screen, MainScreen):
                return

            tree = screen.query_one(ServerTree)
            # Find and select the server node
            for node in tree.root.children:
                if node.data and node.data.server_id == server.id:
                    node.expand()
                    tree.select_node(node)
                    # Show server info
                    for child in node.children:
                        if child.data and child.data.node_type == NodeType.SERVER_INFO:
                            tree.select_node(child)
                            break
                    break
        return callback

    def _ssh_to_server(self, server: Server):
        def callback() -> None:
            ip = server.ip_address
            if not ip:
                self.app.notify("No IP address for this server", severity="error")
                return
            with self.app.suspend():
                subprocess.run(["ssh", "-p", str(server.ssh_port), f"forge@{ip}"])
        return callback
