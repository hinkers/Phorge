"""Site information display panel."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import Static, Button
from textual import work

from rich.markup import escape

from phorge.api.endpoints.sites import SitesAPI
from phorge.widgets.server_tree import NodeData


class SiteInfoPanel(Vertical):
    """Displays detailed site information."""

    DEFAULT_CSS = """
    SiteInfoPanel {
        height: auto;
    }
    SiteInfoPanel .action-bar {
        margin-top: 1;
        height: 3;
        layout: horizontal;
    }
    SiteInfoPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data

    def compose(self) -> ComposeResult:
        name = self.node_data.site_name or self.node_data.label
        yield Static(f"[bold]Site: {name}[/bold]", classes="panel-title")
        yield Static("[dim]Loading...[/dim]", id="site-info-content")
        with Vertical(classes="action-bar"):
            yield Button("SSH", id="btn-ssh", variant="primary")
            yield Button("Open in Browser", id="btn-browser", variant="default")

    def on_mount(self) -> None:
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        try:
            api = SitesAPI(self.app.forge_client)
            site = await api.get(self.node_data.server_id, self.node_data.site_id)

            aliases = ", ".join(site.aliases) if site.aliases else "None"

            info_lines = [
                f"[b]Name:[/b]             {site.name}",
                f"[b]Directory:[/b]        {site.directory or 'N/A'}",
                f"[b]Status:[/b]           {site.status or 'N/A'}",
                f"[b]Project Type:[/b]     {site.project_type or 'N/A'}",
                f"[b]PHP Version:[/b]      {site.php_version or 'N/A'}",
                f"[b]Repository:[/b]       {site.repository or 'N/A'}",
                f"[b]Provider:[/b]         {site.repository_provider or 'N/A'}",
                f"[b]Branch:[/b]           {site.repository_branch or 'N/A'}",
                f"[b]Quick Deploy:[/b]     {'Enabled' if site.quick_deploy else 'Disabled'}",
                f"[b]Wildcards:[/b]        {'Yes' if site.wildcards else 'No'}",
                f"[b]Secured (SSL):[/b]    {'Yes' if site.is_secured else 'No'}",
                f"[b]Aliases:[/b]          {aliases}",
            ]

            content = self.query_one("#site-info-content", Static)
            content.update("\n".join(info_lines))
        except Exception as e:
            content = self.query_one("#site-info-content", Static)
            content.update(f"[red]Error: {escape(str(e))}[/red]")

    async def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-ssh":
            import subprocess

            ip = self.node_data.server_ip
            port = self.node_data.ssh_port
            site_dir = self.node_data.site_directory
            user = self.app.get_ssh_user(self.node_data.server_id)
            if ip:
                cmd = ["ssh", "-t", "-p", str(port), f"{user}@{ip}"]
                if site_dir:
                    cmd.append(f"cd {site_dir} && exec $SHELL -l")
                with self.app.suspend():
                    subprocess.call(cmd)
        elif event.button.id == "btn-browser":
            import webbrowser

            site_name = self.node_data.site_name or self.node_data.label
            webbrowser.open(f"https://{site_name}")
