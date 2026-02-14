"""Daemons panel for managing server daemons."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import DataTable, Static, Button
from textual import work

from phorge.api.endpoints.daemons import DaemonsAPI
from phorge.widgets.server_tree import NodeData


class DaemonsPanel(Vertical):
    """Shows daemons and allows management."""

    DEFAULT_CSS = """
    DaemonsPanel {
        height: 1fr;
    }
    DaemonsPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    DaemonsPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    DaemonsPanel DataTable {
        height: 1fr;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data

    def compose(self) -> ComposeResult:
        yield Static("[bold]Daemons[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Refresh", id="btn-refresh", variant="default")
        yield DataTable(id="daemons-table", cursor_type="row")

    def on_mount(self) -> None:
        table = self.query_one(DataTable)
        table.add_columns("ID", "Command", "User", "Processes", "Status")
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        table = self.query_one(DataTable)
        table.loading = True
        try:
            api = DaemonsAPI(self.app.forge_client)
            daemons = await api.list(self.node_data.server_id)
            table.clear()
            for d in daemons:
                table.add_row(
                    str(d.id),
                    d.command[:50],
                    d.user or "",
                    str(d.processes),
                    d.status or "",
                    key=str(d.id),
                )
            if not daemons:
                self.notify("No daemons found", severity="information")
        except Exception as e:
            self.notify(f"Error: {e}", severity="error", markup=False)
        finally:
            table.loading = False

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-refresh":
            self.load_data()

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        daemon_id = int(str(event.row_key.value))
        self._confirm_restart_daemon(daemon_id)

    @work
    async def _confirm_restart_daemon(self, daemon_id: int) -> None:
        from phorge.screens.confirm import ConfirmModal

        confirmed = await self.app.push_screen_wait(
            ConfirmModal(f"Restart daemon #{daemon_id}?")
        )
        if confirmed:
            api = DaemonsAPI(self.app.forge_client)
            await api.restart(self.node_data.server_id, daemon_id)
            self.notify("Daemon restart initiated")
