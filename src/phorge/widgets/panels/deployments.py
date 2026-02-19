"""Deployments panel with history table and deploy action."""

from __future__ import annotations

import re

from textual.app import ComposeResult
from textual.containers import Vertical, VerticalScroll
from textual.widgets import DataTable, Static, Button
from textual import work

from rich.markup import escape

from phorge.api.endpoints.deployments import DeploymentsAPI
from phorge.widgets.server_tree import NodeData

_ANSI_RE = re.compile(r"\x1b\[[0-9;]*[a-zA-Z]")


def _strip_ansi(text: str) -> str:
    """Remove ANSI escape sequences from text."""
    return _ANSI_RE.sub("", text)


class DeploymentsPanel(Vertical):
    """Shows deployment history for a site."""

    DEFAULT_CSS = """
    DeploymentsPanel {
        height: 1fr;
    }
    DeploymentsPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    DeploymentsPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    DeploymentsPanel DataTable {
        height: 2fr;
    }
    DeploymentsPanel #output-scroll {
        height: 1fr;
        min-height: 6;
        border: solid $primary;
        margin-top: 1;
    }
    DeploymentsPanel #deployment-output {
        padding: 1;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data

    def compose(self) -> ComposeResult:
        yield Static(f"[bold]Deployments - {self.node_data.label}[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Deploy Now", id="btn-deploy", variant="primary")
            yield Button("Reset Status", id="btn-reset", variant="warning")
            yield Button("Refresh", id="btn-refresh", variant="default")
        yield DataTable(id="deployments-table", cursor_type="row")
        with VerticalScroll(id="output-scroll"):
            yield Static("", id="deployment-output")

    def on_mount(self) -> None:
        table = self.query_one(DataTable)
        table.add_columns("ID", "Status", "Commit", "Author", "Message", "Started")
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        table = self.query_one(DataTable)
        table.loading = True
        try:
            api = DeploymentsAPI(self.app.forge_client)
            deployments = await api.list(
                self.node_data.server_id, self.node_data.site_id
            )
            table.clear()
            for d in deployments:
                table.add_row(
                    str(d.id),
                    d.status or "unknown",
                    (d.commit_hash or "")[:8],
                    d.commit_author or "",
                    (d.commit_message or "")[:40],
                    d.started_at or "",
                    key=str(d.id),
                )
        except Exception as e:
            self.notify(f"Error loading deployments: {e}", severity="error", markup=False)
        finally:
            table.loading = False

    @work(exclusive=True, group="output")
    async def load_output(self, deployment_id: int) -> None:
        output_widget = self.query_one("#deployment-output", Static)
        output_widget.update("[dim]Loading output...[/dim]")
        try:
            api = DeploymentsAPI(self.app.forge_client)
            output = await api.get_output(
                self.node_data.server_id, self.node_data.site_id, deployment_id
            )
            clean = escape(_strip_ansi(output)) if output else "[dim]No output[/dim]"
            output_widget.update(f"[bold]Deployment #{deployment_id} Output:[/bold]\n\n{clean}")
        except Exception as e:
            output_widget.update(f"[red]Error: {escape(_strip_ansi(str(e)))}[/red]")

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        deployment_id = int(str(event.row_key.value))
        self.load_output(deployment_id)

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-deploy":
            self._confirm_deploy()
        elif event.button.id == "btn-reset":
            self._reset_status()
        elif event.button.id == "btn-refresh":
            self.load_data()

    @work
    async def _confirm_deploy(self) -> None:
        from phorge.screens.confirm import ConfirmModal

        confirmed = await self.app.push_screen_wait(
            ConfirmModal("Deploy this site now?")
        )
        if confirmed:
            api = DeploymentsAPI(self.app.forge_client)
            await api.deploy(self.node_data.server_id, self.node_data.site_id)
            self.notify("Deployment started")
            self.load_data()

    @work
    async def _reset_status(self) -> None:
        api = DeploymentsAPI(self.app.forge_client)
        await api.reset_status(self.node_data.server_id, self.node_data.site_id)
        self.notify("Deployment status reset")
