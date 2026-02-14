"""SSL certificates panel."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import DataTable, Static, Button
from textual import work

from phorge.api.endpoints.ssl_certificates import SSLCertificatesAPI
from phorge.widgets.server_tree import NodeData


class SSLPanel(Vertical):
    """Shows SSL certificates for a site."""

    DEFAULT_CSS = """
    SSLPanel {
        height: 1fr;
    }
    SSLPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    SSLPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    SSLPanel DataTable {
        height: 1fr;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data

    def compose(self) -> ComposeResult:
        yield Static("[bold]SSL Certificates[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Let's Encrypt", id="btn-letsencrypt", variant="primary")
            yield Button("Refresh", id="btn-refresh", variant="default")
        yield DataTable(id="ssl-table", cursor_type="row")

    def on_mount(self) -> None:
        table = self.query_one(DataTable)
        table.add_columns("ID", "Domain", "Type", "Active", "Status")
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        table = self.query_one(DataTable)
        table.loading = True
        try:
            api = SSLCertificatesAPI(self.app.forge_client)
            certs = await api.list(
                self.node_data.server_id, self.node_data.site_id
            )
            table.clear()
            for c in certs:
                table.add_row(
                    str(c.id),
                    c.domain or "",
                    c.type or "",
                    "Yes" if c.active else "No",
                    c.status or "",
                    key=str(c.id),
                )
            if not certs:
                self.notify("No SSL certificates found", severity="information")
        except Exception as e:
            self.notify(f"Error: {e}", severity="error", markup=False)
        finally:
            table.loading = False

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-refresh":
            self.load_data()
        elif event.button.id == "btn-letsencrypt":
            self._request_letsencrypt()

    @work
    async def _request_letsencrypt(self) -> None:
        from phorge.screens.input_modal import InputModal

        domain = await self.app.push_screen_wait(
            InputModal(
                "Let's Encrypt Domain",
                placeholder="example.com",
            )
        )
        if domain:
            api = SSLCertificatesAPI(self.app.forge_client)
            await api.obtain_letsencrypt(
                self.node_data.server_id,
                self.node_data.site_id,
                [domain],
            )
            self.notify("Let's Encrypt certificate requested")
            self.load_data()
