"""Deployment script viewer/editor panel."""

from __future__ import annotations

import subprocess
import tempfile
from pathlib import Path

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import Static, Button
from textual import work

from rich.markup import escape

from phorge.api.endpoints.deployments import DeploymentsAPI
from phorge.config import load_config
from phorge.widgets.server_tree import NodeData


class DeploymentScriptPanel(Vertical):
    """Shows the deployment script and allows editing in external editor."""

    DEFAULT_CSS = """
    DeploymentScriptPanel {
        height: 1fr;
    }
    DeploymentScriptPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    DeploymentScriptPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    DeploymentScriptPanel #script-content {
        height: 1fr;
        border: solid $primary;
        padding: 1;
        overflow-y: auto;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data
        self._script_content = ""

    def compose(self) -> ComposeResult:
        yield Static("[bold]Deployment Script[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Edit in Editor", id="btn-edit", variant="primary")
        yield Static("[dim]Loading...[/dim]", id="script-content")

    def on_mount(self) -> None:
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        try:
            api = DeploymentsAPI(self.app.forge_client)
            self._script_content = await api.get_script(
                self.node_data.server_id, self.node_data.site_id
            )
            content = self.query_one("#script-content", Static)
            content.update(self._script_content or "[dim]No deployment script[/dim]")
        except Exception as e:
            content = self.query_one("#script-content", Static)
            content.update(f"[red]Error: {escape(str(e))}[/red]")

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-edit":
            self._edit_in_editor()

    @work
    async def _edit_in_editor(self) -> None:
        config = load_config()
        editor_cmd = config.editor.command

        with tempfile.NamedTemporaryFile(
            mode="w", suffix=".sh", delete=False, prefix="phorge_deploy_"
        ) as f:
            f.write(self._script_content)
            tmp_path = Path(f.name)

        try:
            mtime_before = tmp_path.stat().st_mtime

            with self.app.suspend():
                subprocess.run(
                    [editor_cmd, "--wait", str(tmp_path)] if editor_cmd == "code"
                    else [editor_cmd, str(tmp_path)],
                    shell=True,
                )

            mtime_after = tmp_path.stat().st_mtime
            if mtime_after != mtime_before:
                new_content = tmp_path.read_text()
                from phorge.screens.confirm import ConfirmModal

                confirmed = await self.app.push_screen_wait(
                    ConfirmModal("Upload modified deployment script?")
                )
                if confirmed:
                    api = DeploymentsAPI(self.app.forge_client)
                    await api.update_script(
                        self.node_data.server_id,
                        self.node_data.site_id,
                        new_content,
                    )
                    self._script_content = new_content
                    content = self.query_one("#script-content", Static)
                    content.update(new_content)
                    self.notify("Deployment script updated")
        finally:
            tmp_path.unlink(missing_ok=True)
