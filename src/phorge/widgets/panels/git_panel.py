"""Git repository information panel."""

from __future__ import annotations

import webbrowser

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import Static, Button
from textual import work

from rich.markup import escape

from phorge.api.endpoints.sites import SitesAPI
from phorge.widgets.server_tree import NodeData


class GitPanel(Vertical):
    """Shows git repository info and allows opening in browser."""

    DEFAULT_CSS = """
    GitPanel {
        height: auto;
    }
    GitPanel .action-bar {
        margin-top: 1;
        height: 3;
        layout: horizontal;
    }
    GitPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data
        self._repo_url: str | None = None

    def compose(self) -> ComposeResult:
        yield Static("[bold]Git Repository[/bold]", classes="panel-title")
        yield Static("[dim]Loading...[/dim]", id="git-info")
        with Vertical(classes="action-bar"):
            yield Button("Open in Browser", id="btn-open", variant="primary")
            yield Button("Refresh", id="btn-refresh", variant="default")

    def on_mount(self) -> None:
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        try:
            api = SitesAPI(self.app.forge_client)
            site = await api.get(self.node_data.server_id, self.node_data.site_id)

            self._repo_url = self._build_repo_url(
                site.repository_provider, site.repository
            )

            info_lines = [
                f"[b]Repository:[/b]     {site.repository or 'N/A'}",
                f"[b]Provider:[/b]       {site.repository_provider or 'N/A'}",
                f"[b]Branch:[/b]         {site.repository_branch or 'N/A'}",
                f"[b]Status:[/b]         {site.repository_status or 'N/A'}",
                f"[b]Quick Deploy:[/b]   {'Enabled' if site.quick_deploy else 'Disabled'}",
            ]

            if self._repo_url:
                info_lines.append(f"[b]URL:[/b]            {self._repo_url}")

            content = self.query_one("#git-info", Static)
            content.update("\n".join(info_lines))
        except Exception as e:
            content = self.query_one("#git-info", Static)
            content.update(f"[red]Error: {escape(str(e))}[/red]")

    def _build_repo_url(
        self, provider: str | None, repository: str | None
    ) -> str | None:
        if not repository:
            return None
        if provider == "github":
            return f"https://github.com/{repository}"
        if provider == "gitlab":
            return f"https://gitlab.com/{repository}"
        if provider == "bitbucket":
            return f"https://bitbucket.org/{repository}"
        return None

    async def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-open":
            if self._repo_url:
                webbrowser.open(self._repo_url)
            else:
                self.notify("No repository URL available", severity="warning")
        elif event.button.id == "btn-refresh":
            self.load_data()
