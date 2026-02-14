"""Workers panel for managing queue workers."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import DataTable, Static, Button
from textual import work

from phorge.api.endpoints.workers import WorkersAPI
from phorge.widgets.server_tree import NodeData


class WorkersPanel(Vertical):
    """Shows queue workers for a site."""

    DEFAULT_CSS = """
    WorkersPanel {
        height: 1fr;
    }
    WorkersPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    WorkersPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    WorkersPanel DataTable {
        height: 1fr;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data

    def compose(self) -> ComposeResult:
        yield Static("[bold]Workers[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Refresh", id="btn-refresh", variant="default")
        yield DataTable(id="workers-table", cursor_type="row")

    def on_mount(self) -> None:
        table = self.query_one(DataTable)
        table.add_columns("ID", "Connection", "Queue", "Processes", "Timeout", "Status")
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        table = self.query_one(DataTable)
        table.loading = True
        try:
            api = WorkersAPI(self.app.forge_client)
            workers = await api.list(
                self.node_data.server_id, self.node_data.site_id
            )
            table.clear()
            for w in workers:
                table.add_row(
                    str(w.id),
                    w.connection or "",
                    w.queue or "",
                    str(w.processes),
                    str(w.timeout),
                    w.status or "",
                    key=str(w.id),
                )
            if not workers:
                self.notify("No workers found", severity="information")
        except Exception as e:
            self.notify(f"Error: {e}", severity="error", markup=False)
        finally:
            table.loading = False

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-refresh":
            self.load_data()

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        worker_id = int(str(event.row_key.value))
        self._confirm_restart_worker(worker_id)

    @work
    async def _confirm_restart_worker(self, worker_id: int) -> None:
        from phorge.screens.confirm import ConfirmModal

        confirmed = await self.app.push_screen_wait(
            ConfirmModal(f"Restart worker #{worker_id}?")
        )
        if confirmed:
            api = WorkersAPI(self.app.forge_client)
            await api.restart(
                self.node_data.server_id, self.node_data.site_id, worker_id
            )
            self.notify("Worker restart initiated")
