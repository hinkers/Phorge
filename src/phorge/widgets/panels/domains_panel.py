"""Domains/aliases panel for managing site domains."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import DataTable, Static, Button
from textual import work

from phorge.api.endpoints.sites import SitesAPI
from phorge.widgets.server_tree import NodeData


class DomainsPanel(Vertical):
    """Shows and manages site aliases/domains."""

    DEFAULT_CSS = """
    DomainsPanel {
        height: 1fr;
    }
    DomainsPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    DomainsPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    DomainsPanel DataTable {
        height: 1fr;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data
        self._aliases: list[str] = []

    def compose(self) -> ComposeResult:
        yield Static("[bold]Domains / Aliases[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Add Domain", id="btn-add", variant="primary")
            yield Button("Refresh", id="btn-refresh", variant="default")
        yield DataTable(id="domains-table", cursor_type="row")

    def on_mount(self) -> None:
        table = self.query_one(DataTable)
        table.add_columns("#", "Domain")
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        table = self.query_one(DataTable)
        table.loading = True
        try:
            api = SitesAPI(self.app.forge_client)
            site = await api.get(self.node_data.server_id, self.node_data.site_id)
            self._aliases = site.aliases or []

            table.clear()
            # Primary domain
            table.add_row("1", site.name, key="primary")
            # Aliases
            for i, alias in enumerate(self._aliases):
                table.add_row(str(i + 2), alias, key=f"alias-{i}")
        except Exception as e:
            self.notify(f"Error: {e}", severity="error", markup=False)
        finally:
            table.loading = False

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-refresh":
            self.load_data()
        elif event.button.id == "btn-add":
            self._add_domain()

    @work
    async def _add_domain(self) -> None:
        from phorge.screens.input_modal import InputModal

        domain = await self.app.push_screen_wait(
            InputModal("Add Domain Alias", placeholder="example.com")
        )
        if domain:
            self._aliases.append(domain)
            api = SitesAPI(self.app.forge_client)
            await api.update_aliases(
                self.node_data.server_id,
                self.node_data.site_id,
                self._aliases,
            )
            self.notify(f"Added alias: {domain}")
            self.load_data()
