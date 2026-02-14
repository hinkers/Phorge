"""Database users panel."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import DataTable, Static, Button
from textual import work

from phorge.api.endpoints.databases import DatabaseUsersAPI
from phorge.widgets.server_tree import NodeData


class DatabaseUsersPanel(Vertical):
    """Shows database users and allows management."""

    DEFAULT_CSS = """
    DatabaseUsersPanel {
        height: 1fr;
    }
    DatabaseUsersPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    DatabaseUsersPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    DatabaseUsersPanel DataTable {
        height: 1fr;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data

    def compose(self) -> ComposeResult:
        yield Static("[bold]Database Users[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Refresh", id="btn-refresh", variant="default")
        yield DataTable(id="db-users-table", cursor_type="row")

    def on_mount(self) -> None:
        table = self.query_one(DataTable)
        table.add_columns("ID", "Name", "Status", "Databases")
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        table = self.query_one(DataTable)
        table.loading = True
        try:
            api = DatabaseUsersAPI(self.app.forge_client)
            users = await api.list(self.node_data.server_id)
            table.clear()
            for u in users:
                db_ids = ", ".join(str(d) for d in u.databases) if u.databases else "None"
                table.add_row(
                    str(u.id),
                    u.name,
                    u.status or "",
                    db_ids,
                    key=str(u.id),
                )
            if not users:
                self.notify("No database users found", severity="information")
        except Exception as e:
            self.notify(f"Error: {e}", severity="error", markup=False)
        finally:
            table.loading = False

    async def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-refresh":
            self.load_data()
