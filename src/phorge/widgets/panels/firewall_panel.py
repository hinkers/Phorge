"""Firewall rules panel."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import DataTable, Static, Button
from textual import work

from phorge.api.endpoints.firewall import FirewallAPI
from phorge.widgets.server_tree import NodeData


class FirewallPanel(Vertical):
    """Shows firewall rules and allows management."""

    DEFAULT_CSS = """
    FirewallPanel {
        height: 1fr;
    }
    FirewallPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    FirewallPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    FirewallPanel DataTable {
        height: 1fr;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data

    def compose(self) -> ComposeResult:
        yield Static("[bold]Firewall Rules[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Refresh", id="btn-refresh", variant="default")
        yield DataTable(id="firewall-table", cursor_type="row")

    def on_mount(self) -> None:
        table = self.query_one(DataTable)
        table.add_columns("ID", "Name", "Port", "IP Address", "Type", "Status")
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        table = self.query_one(DataTable)
        table.loading = True
        try:
            api = FirewallAPI(self.app.forge_client)
            rules = await api.list(self.node_data.server_id)
            table.clear()
            for r in rules:
                table.add_row(
                    str(r.id),
                    r.name,
                    str(r.port) if r.port else "",
                    r.ip_address or "Any",
                    r.type or "",
                    r.status or "",
                    key=str(r.id),
                )
            if not rules:
                self.notify("No firewall rules found", severity="information")
        except Exception as e:
            self.notify(f"Error: {e}", severity="error", markup=False)
        finally:
            table.loading = False

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-refresh":
            self.load_data()

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        rule_id = int(str(event.row_key.value))
        self._confirm_delete_rule(rule_id)

    @work
    async def _confirm_delete_rule(self, rule_id: int) -> None:
        from phorge.screens.confirm import ConfirmModal

        confirmed = await self.app.push_screen_wait(
            ConfirmModal(f"Delete firewall rule #{rule_id}?")
        )
        if confirmed:
            api = FirewallAPI(self.app.forge_client)
            await api.delete(self.node_data.server_id, rule_id)
            self.notify("Firewall rule deleted")
            self.load_data()
