"""Server information display panel."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import Static, Button
from textual import work

from rich.markup import escape

from phorge.api.endpoints.servers import ServersAPI
from phorge.widgets.server_tree import NodeData


class ServerInfoPanel(Vertical):
    """Displays detailed server information."""

    DEFAULT_CSS = """
    ServerInfoPanel {
        height: auto;
    }
    ServerInfoPanel .info-grid {
        padding: 0 1;
    }
    ServerInfoPanel .info-row {
        height: 1;
        margin-bottom: 0;
    }
    ServerInfoPanel .info-label {
        width: 22;
        color: $text-muted;
    }
    ServerInfoPanel .info-value {
        width: 1fr;
    }
    ServerInfoPanel .action-bar {
        margin-top: 1;
        height: 3;
        layout: horizontal;
    }
    ServerInfoPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data

    def compose(self) -> ComposeResult:
        yield Static(f"[bold]Server: {self.node_data.label}[/bold]", classes="panel-title")
        yield Static("[dim]Loading...[/dim]", id="server-info-content")
        with Vertical(classes="action-bar"):
            yield Button("SSH", id="btn-ssh", variant="primary")
            yield Button("Files", id="btn-sftp", variant="default")
            yield Button("Reboot", id="btn-reboot", variant="error")

    def on_mount(self) -> None:
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        try:
            api = ServersAPI(self.app.forge_client)
            server = await api.get(self.node_data.server_id)

            info_lines = [
                f"[b]Name:[/b]           {server.name}",
                f"[b]IP Address:[/b]     {server.ip_address or 'N/A'}",
                f"[b]Private IP:[/b]     {server.private_ip_address or 'N/A'}",
                f"[b]Region:[/b]         {server.region or 'N/A'}",
                f"[b]Provider:[/b]       {server.provider or 'N/A'}",
                f"[b]Type:[/b]           {server.type or 'N/A'}",
                f"[b]Status:[/b]         {server.status or 'N/A'}",
                f"[b]PHP Version:[/b]    {server.php_version or 'N/A'}",
                f"[b]PHP CLI:[/b]        {server.php_cli_version or 'N/A'}",
                f"[b]Ubuntu:[/b]         {server.ubuntu_version or 'N/A'}",
                f"[b]Database:[/b]       {server.database_type or 'N/A'}",
                f"[b]DB Status:[/b]      {server.db_status or 'N/A'}",
                f"[b]Redis Status:[/b]   {server.redis_status or 'N/A'}",
                f"[b]SSH Port:[/b]       {server.ssh_port}",
                f"[b]Ready:[/b]          {'Yes' if server.is_ready else 'No'}",
            ]

            content = self.query_one("#server-info-content", Static)
            content.update("\n".join(info_lines))
        except Exception as e:
            content = self.query_one("#server-info-content", Static)
            content.update(f"[red]Error: {escape(str(e))}[/red]")

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-ssh":
            self._ssh_to_server()
        elif event.button.id == "btn-sftp":
            self._sftp_to_server()
        elif event.button.id == "btn-reboot":
            self._confirm_reboot()

    def _ssh_to_server(self) -> None:
        import subprocess

        ip = self.node_data.server_ip
        port = self.node_data.ssh_port
        user = self.app.get_ssh_user(self.node_data.server_id)
        if ip:
            with self.app.suspend():
                subprocess.call(["ssh", "-t", "-p", str(port), f"{user}@{ip}"])

    def _sftp_to_server(self) -> None:
        import subprocess

        ip = self.node_data.server_ip
        port = self.node_data.ssh_port
        user = self.app.get_ssh_user(self.node_data.server_id)
        if ip:
            address = f"sftp://{user}@{ip}:{port}:/home/{user}"
            with self.app.suspend():
                subprocess.call(["termscp", address])

    @work
    async def _confirm_reboot(self) -> None:
        from phorge.screens.confirm import ConfirmModal

        confirmed = await self.app.push_screen_wait(
            ConfirmModal("Are you sure you want to reboot this server?")
        )
        if confirmed:
            api = ServersAPI(self.app.forge_client)
            await api.reboot(self.node_data.server_id)
            self.notify("Server reboot initiated")
