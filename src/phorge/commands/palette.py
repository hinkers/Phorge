"""Command palette providers for quick navigation."""

from __future__ import annotations

import subprocess

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
        yield DiscoveryHit(
            "Switch Server",
            self._switch_server,
            help="Open server picker to select a different server",
        )

        yield DiscoveryHit(
            "Edit Configuration",
            self._open_config,
            help="Edit API key, editor, and UI settings",
        )

        for server in self._servers:
            ip = server.ip_address or "no ip"
            yield DiscoveryHit(
                f"SSH to {server.name}",
                self._ssh_to_server(server),
                help=f"Connect via SSH to {ip}",
            )

    async def search(self, query: str) -> Hits:
        """Fuzzy search across servers and actions."""
        matcher = self.matcher(query)

        # Switch Server
        switch_label = "Switch Server"
        score = matcher.match(switch_label)
        if score > 0:
            yield Hit(score, matcher.highlight(switch_label), self._switch_server)

        # Config
        config_label = "Edit Configuration"
        score = matcher.match(config_label)
        if score > 0:
            yield Hit(score, matcher.highlight(config_label), self._open_config)

        # Refresh
        refresh_label = "Refresh Server"
        score = matcher.match(refresh_label)
        if score > 0:
            yield Hit(score, matcher.highlight(refresh_label), self._refresh)

        for server in self._servers:
            ip = server.ip_address or ""

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

    def _switch_server(self) -> None:
        from phorge.screens.main import MainScreen
        screen = self.app.screen
        if isinstance(screen, MainScreen):
            screen.action_switch_server()

    def _ssh_to_server(self, server: Server):
        def callback() -> None:
            ip = server.ip_address
            if not ip:
                self.app.notify("No IP address for this server", severity="error")
                return
            with self.app.suspend():
                subprocess.call(["ssh", "-t", "-p", str(server.ssh_port), f"forge@{ip}"])
        return callback
