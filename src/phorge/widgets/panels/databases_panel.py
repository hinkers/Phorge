"""Databases panel for managing server databases."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import DataTable, Static, Button
from textual import work

from phorge.api.endpoints.databases import DatabasesAPI
from phorge.widgets.server_tree import NodeData


class DatabasesPanel(Vertical):
    """Shows databases and allows creation/deletion."""

    DEFAULT_CSS = """
    DatabasesPanel {
        height: 1fr;
    }
    DatabasesPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    DatabasesPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    DatabasesPanel DataTable {
        height: 1fr;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data

    def compose(self) -> ComposeResult:
        yield Static("[bold]Databases[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Create Database", id="btn-create", variant="primary")
            yield Button("Sync", id="btn-sync", variant="default")
            yield Button("Refresh", id="btn-refresh", variant="default")
        yield DataTable(id="databases-table", cursor_type="row")

    def on_mount(self) -> None:
        table = self.query_one(DataTable)
        table.add_columns("ID", "Name", "Status", "Synced")
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        table = self.query_one(DataTable)
        table.loading = True
        try:
            api = DatabasesAPI(self.app.forge_client)
            databases = await api.list(self.node_data.server_id)
            table.clear()
            for db in databases:
                table.add_row(
                    str(db.id),
                    db.name,
                    db.status or "",
                    "Yes" if db.is_synced else "No",
                    key=str(db.id),
                )
            if not databases:
                self.notify("No databases found", severity="information")
        except Exception as e:
            self.notify(f"Error: {e}", severity="error", markup=False)
        finally:
            table.loading = False

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-refresh":
            self.load_data()
        elif event.button.id == "btn-sync":
            self._sync_databases()
        elif event.button.id == "btn-create":
            self._create_database()

    @work
    async def _sync_databases(self) -> None:
        api = DatabasesAPI(self.app.forge_client)
        await api.sync(self.node_data.server_id)
        self.notify("Database sync initiated")
        self.load_data()

    @work
    async def _create_database(self) -> None:
        from phorge.screens.input_modal import InputModal

        name = await self.app.push_screen_wait(
            InputModal("Database Name", placeholder="my_database")
        )
        if name:
            api = DatabasesAPI(self.app.forge_client)
            await api.create(self.node_data.server_id, name)
            self.notify(f"Database '{name}' created")
            self.load_data()

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        db_id = int(str(event.row_key.value))
        self._confirm_delete_database(db_id)

    @work
    async def _confirm_delete_database(self, db_id: int) -> None:
        from phorge.screens.confirm import ConfirmModal

        confirmed = await self.app.push_screen_wait(
            ConfirmModal(f"Delete database #{db_id}? This cannot be undone.")
        )
        if confirmed:
            api = DatabasesAPI(self.app.forge_client)
            await api.delete(self.node_data.server_id, db_id)
            self.notify("Database deleted")
            self.load_data()
