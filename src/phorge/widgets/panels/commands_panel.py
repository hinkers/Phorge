"""Commands panel for executing and viewing site commands."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import DataTable, Static, Button
from textual import work

from phorge.api.endpoints.commands import CommandsAPI
from phorge.widgets.server_tree import NodeData


class CommandsPanel(Vertical):
    """Shows command history and allows executing new commands."""

    DEFAULT_CSS = """
    CommandsPanel {
        height: 1fr;
    }
    CommandsPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    CommandsPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    CommandsPanel DataTable {
        height: 1fr;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data

    def compose(self) -> ComposeResult:
        yield Static("[bold]Commands[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Execute Command", id="btn-execute", variant="primary")
            yield Button("Refresh", id="btn-refresh", variant="default")
        yield DataTable(id="commands-table", cursor_type="row")

    def on_mount(self) -> None:
        table = self.query_one(DataTable)
        table.add_columns("ID", "Command", "Status", "User", "Created", "Duration")
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        table = self.query_one(DataTable)
        table.loading = True
        try:
            api = CommandsAPI(self.app.forge_client)
            commands = await api.list(
                self.node_data.server_id, self.node_data.site_id
            )
            table.clear()
            for c in commands:
                table.add_row(
                    str(c.id),
                    c.command[:50],
                    c.status or "",
                    c.user_name or "",
                    c.created_at or "",
                    f"{c.duration}s" if c.duration else "",
                    key=str(c.id),
                )
        except Exception as e:
            self.notify(f"Error: {e}", severity="error", markup=False)
        finally:
            table.loading = False

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-refresh":
            self.load_data()
        elif event.button.id == "btn-execute":
            self._execute_command()

    @work
    async def _execute_command(self) -> None:
        from phorge.screens.input_modal import InputModal

        command = await self.app.push_screen_wait(
            InputModal("Execute Command", placeholder="php artisan migrate")
        )
        if command:
            api = CommandsAPI(self.app.forge_client)
            await api.execute(
                self.node_data.server_id, self.node_data.site_id, command
            )
            self.notify("Command executed")
            self.load_data()
