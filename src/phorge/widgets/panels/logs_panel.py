"""Log viewer panel for server and site logs."""

from __future__ import annotations

import subprocess
import tempfile
from pathlib import Path

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import Static, Button
from textual import work

from rich.markup import escape

from phorge.api.endpoints.logs import LogsAPI
from phorge.config import load_config
from phorge.widgets.server_tree import NodeData


class LogsPanel(Vertical):
    """Displays server or site logs."""

    DEFAULT_CSS = """
    LogsPanel {
        height: 1fr;
    }
    LogsPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    LogsPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    LogsPanel #log-content {
        height: 1fr;
        border: solid $primary;
        padding: 1;
        overflow-y: auto;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data
        self._log_content = ""

    def compose(self) -> ComposeResult:
        yield Static("[bold]Logs[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Open in Editor", id="btn-edit", variant="primary")
            yield Button("Refresh", id="btn-refresh", variant="default")
            if self.node_data.site_id is not None:
                yield Button("Clear Logs", id="btn-clear", variant="error")
        yield Static("[dim]Loading...[/dim]", id="log-content")

    def on_mount(self) -> None:
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        try:
            api = LogsAPI(self.app.forge_client)
            if self.node_data.site_id is not None:
                self._log_content = await api.get_site_logs(
                    self.node_data.server_id, self.node_data.site_id
                )
            else:
                self._log_content = await api.get_server_logs(self.node_data.server_id)

            content = self.query_one("#log-content", Static)
            content.update(self._log_content or "[dim]No logs available[/dim]")
        except Exception as e:
            content = self.query_one("#log-content", Static)
            content.update(f"[red]Error: {escape(str(e))}[/red]")

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-edit":
            self._open_in_editor()
        elif event.button.id == "btn-refresh":
            self.load_data()
        elif event.button.id == "btn-clear":
            self._confirm_clear_logs()

    @work
    async def _open_in_editor(self) -> None:
        if not self._log_content:
            self.notify("No logs to open", severity="warning")
            return

        config = load_config()
        editor_cmd = config.editor.command
        site_name = self.node_data.site_name or "server"

        with tempfile.NamedTemporaryFile(
            mode="w", suffix=".log", delete=False, prefix=f"phorge_{site_name}_"
        ) as f:
            f.write(self._log_content)
            tmp_path = Path(f.name)

        try:
            with self.app.suspend():
                subprocess.run(
                    [editor_cmd, "--wait", str(tmp_path)] if editor_cmd == "code"
                    else [editor_cmd, str(tmp_path)],
                    shell=True,
                )
        finally:
            tmp_path.unlink(missing_ok=True)

    @work
    async def _confirm_clear_logs(self) -> None:
        from phorge.screens.confirm import ConfirmModal

        confirmed = await self.app.push_screen_wait(
            ConfirmModal("Clear all site logs?")
        )
        if confirmed and self.node_data.site_id is not None:
            api = LogsAPI(self.app.forge_client)
            await api.clear_site_logs(
                self.node_data.server_id, self.node_data.site_id
            )
            self.notify("Logs cleared")
            self.load_data()
