"""Environment file viewer/editor panel."""

from __future__ import annotations

import asyncio
import os
import subprocess
import tempfile
from pathlib import Path

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import Static, Button
from textual import work

from rich.markup import escape

from phorge.api.endpoints.environment import EnvironmentAPI
from phorge.config import load_config
from phorge.widgets.server_tree import NodeData


class EnvironmentPanel(Vertical):
    """Shows .env file contents and allows editing in external editor."""

    DEFAULT_CSS = """
    EnvironmentPanel {
        height: 1fr;
    }
    EnvironmentPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    EnvironmentPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    EnvironmentPanel #env-content {
        height: 1fr;
        border: solid $primary;
        padding: 1;
        overflow-y: auto;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data
        self._env_content = ""
        self._data_loaded = asyncio.Event()

    def compose(self) -> ComposeResult:
        yield Static("[bold]Environment File[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Open in Editor", id="btn-edit", variant="primary")
            yield Button("Refresh", id="btn-refresh", variant="default")
        yield Static("[dim]Loading...[/dim]", id="env-content")

    def on_mount(self) -> None:
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        try:
            api = EnvironmentAPI(self.app.forge_client)
            self._env_content = await api.get(
                self.node_data.server_id, self.node_data.site_id
            )
            content = self.query_one("#env-content", Static)
            content.update(self._env_content or "[dim]Empty .env file[/dim]")
        except Exception as e:
            content = self.query_one("#env-content", Static)
            content.update(f"[red]Error: {escape(str(e))}[/red]")
        finally:
            self._data_loaded.set()

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-edit":
            self._edit_in_editor()
        elif event.button.id == "btn-refresh":
            self.load_data()

    @work
    async def _edit_in_editor(self) -> None:
        await self._data_loaded.wait()
        config = load_config()
        editor_cmd = config.editor.command

        with tempfile.NamedTemporaryFile(
            mode="w", suffix=".env", delete=False, prefix="phorge_"
        ) as f:
            f.write(self._env_content)
            tmp_path = Path(f.name)
        os.chmod(f.name, 0o600)

        try:
            mtime_before = tmp_path.stat().st_mtime

            with self.app.suspend():
                subprocess.run(
                    [editor_cmd, "--wait", str(tmp_path)] if editor_cmd == "code"
                    else [editor_cmd, str(tmp_path)],
                )

            mtime_after = tmp_path.stat().st_mtime
            if mtime_after != mtime_before:
                new_content = tmp_path.read_text()
                from phorge.screens.confirm import ConfirmModal

                confirmed = await self.app.push_screen_wait(
                    ConfirmModal("Upload modified .env file?")
                )
                if confirmed:
                    api = EnvironmentAPI(self.app.forge_client)
                    await api.update(
                        self.node_data.server_id,
                        self.node_data.site_id,
                        new_content,
                    )
                    self._env_content = new_content
                    content = self.query_one("#env-content", Static)
                    content.update(new_content)
                    self.notify("Environment file updated")
        finally:
            tmp_path.unlink(missing_ok=True)
