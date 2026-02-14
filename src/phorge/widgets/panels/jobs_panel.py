"""Scheduled jobs panel."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import DataTable, Static, Button
from textual import work

from phorge.api.endpoints.jobs import ScheduledJobsAPI
from phorge.widgets.server_tree import NodeData


class ScheduledJobsPanel(Vertical):
    """Shows scheduled jobs and allows management."""

    DEFAULT_CSS = """
    ScheduledJobsPanel {
        height: 1fr;
    }
    ScheduledJobsPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    ScheduledJobsPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    ScheduledJobsPanel DataTable {
        height: 1fr;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data

    def compose(self) -> ComposeResult:
        yield Static("[bold]Scheduled Jobs[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Refresh", id="btn-refresh", variant="default")
        yield DataTable(id="jobs-table", cursor_type="row")

    def on_mount(self) -> None:
        table = self.query_one(DataTable)
        table.add_columns("ID", "Command", "User", "Frequency", "Cron", "Status")
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        table = self.query_one(DataTable)
        table.loading = True
        try:
            api = ScheduledJobsAPI(self.app.forge_client)
            jobs = await api.list(self.node_data.server_id)
            table.clear()
            for j in jobs:
                table.add_row(
                    str(j.id),
                    j.command[:50],
                    j.user or "",
                    j.frequency or "",
                    j.cron or "",
                    j.status or "",
                    key=str(j.id),
                )
            if not jobs:
                self.notify("No scheduled jobs found", severity="information")
        except Exception as e:
            self.notify(f"Error: {e}", severity="error", markup=False)
        finally:
            table.loading = False

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-refresh":
            self.load_data()

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        job_id = int(str(event.row_key.value))
        self._confirm_delete_job(job_id)

    @work
    async def _confirm_delete_job(self, job_id: int) -> None:
        from phorge.screens.confirm import ConfirmModal

        confirmed = await self.app.push_screen_wait(
            ConfirmModal(f"Delete scheduled job #{job_id}?")
        )
        if confirmed:
            api = ScheduledJobsAPI(self.app.forge_client)
            await api.delete(self.node_data.server_id, job_id)
            self.notify("Scheduled job deleted")
            self.load_data()
