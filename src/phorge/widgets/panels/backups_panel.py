"""Backups panel for managing backup configurations and viewing backup history."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import DataTable, Static, Button
from textual import work

from phorge.api.endpoints.backups import BackupsAPI
from phorge.api.models import BackupConfig
from phorge.widgets.server_tree import NodeData


class BackupsPanel(Vertical):
    """Shows backup configurations and their backup history."""

    DEFAULT_CSS = """
    BackupsPanel {
        height: 1fr;
    }
    BackupsPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    BackupsPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    BackupsPanel DataTable {
        height: 1fr;
    }
    BackupsPanel #backups-history-table {
        height: 1fr;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data
        self._configs: list[BackupConfig] = []

    def compose(self) -> ComposeResult:
        yield Static("[bold]Backup Configurations[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Refresh", id="btn-refresh", variant="default")
        yield DataTable(id="backups-table", cursor_type="row")
        yield Static("[bold]Backup History[/bold]  [dim](select a config above)[/dim]", id="history-title")
        yield DataTable(id="backups-history-table", cursor_type="row")

    def on_mount(self) -> None:
        configs_table = self.query_one("#backups-table", DataTable)
        configs_table.add_columns("ID", "Provider", "Frequency", "Time", "Day")

        history_table = self.query_one("#backups-history-table", DataTable)
        history_table.add_columns("ID", "Status", "Date", "Size", "Duration")

        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        configs_table = self.query_one("#backups-table", DataTable)
        configs_table.loading = True
        try:
            api = BackupsAPI(self.app.forge_client)
            self._configs = await api.list_configs(self.node_data.server_id)
            configs_table.clear()
            for c in self._configs:
                configs_table.add_row(
                    str(c.id),
                    c.provider or "",
                    c.frequency or "",
                    c.backup_time or c.time or "",
                    str(c.day_of_week) if c.day_of_week is not None else "",
                    key=str(c.id),
                )

            # If there are configs, show backups from the first one
            if self._configs:
                self._show_backups_for_config(self._configs[0])
            else:
                self.notify("No backup configurations found", severity="information")
        except Exception as e:
            self.notify(f"Error: {e}", severity="error", markup=False)
        finally:
            configs_table.loading = False

    def _show_backups_for_config(self, config: BackupConfig) -> None:
        """Display backup entries for the given config."""
        history_table = self.query_one("#backups-history-table", DataTable)
        history_title = self.query_one("#history-title", Static)
        history_title.update(
            f"[bold]Backup History[/bold]  [dim](Config #{config.id} — {config.provider or 'unknown'})[/dim]"
        )
        history_table.clear()

        backups = config.backups or []
        for b in backups:
            history_table.add_row(
                str(b.id),
                b.status or "",
                b.date or "",
                str(b.size) if b.size is not None else "",
                str(b.duration) if b.duration is not None else "",
                key=str(b.id),
            )

        if not backups:
            history_title.update(
                f"[bold]Backup History[/bold]  [dim](Config #{config.id} — no backups yet)[/dim]"
            )

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-refresh":
            self.load_data()

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        if event.data_table.id == "backups-table":
            config_id = int(str(event.row_key.value))
            # Find the config and show its backups
            for c in self._configs:
                if c.id == config_id:
                    self._show_backups_for_config(c)
                    return
            # If not found in cache, offer to run backup
            self._confirm_run_backup(config_id)
        elif event.data_table.id == "backups-history-table":
            # Could add restore/delete actions here
            pass

    @work
    async def _confirm_run_backup(self, config_id: int) -> None:
        from phorge.screens.confirm import ConfirmModal

        confirmed = await self.app.push_screen_wait(
            ConfirmModal(f"Run backup for configuration #{config_id}?")
        )
        if confirmed:
            api = BackupsAPI(self.app.forge_client)
            await api.run_backup(self.node_data.server_id, config_id)
            self.notify("Backup started")
            self.load_data()
