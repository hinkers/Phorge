"""Main application screen with two-pane layout."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Horizontal
from textual.screen import Screen
from textual.widgets import Header, Footer
from textual import work

from phorge.api.endpoints.sites import SitesAPI
from phorge.api.exceptions import ForgeAPIError
from phorge.api.models import Server
from phorge.config import load_config
from phorge.screens.server_picker import ServerPicker
from phorge.widgets.detail_panel import DetailPanel
from phorge.widgets.server_tree import ServerTree, NodeData, NodeType


class MainScreen(Screen):
    """Primary application screen with tree navigation and detail panel."""

    DEFAULT_CSS = """
    MainScreen {
        layout: vertical;
    }
    #main-container {
        width: 100%;
        height: 1fr;
    }
    """

    BINDINGS = [
        ("ctrl+r", "refresh", "Refresh"),
        ("ctrl+g", "switch_server", "Servers"),
    ]

    def __init__(self, **kwargs) -> None:
        super().__init__(**kwargs)
        self._current_server: Server | None = None

    def compose(self) -> ComposeResult:
        yield Header()
        with Horizontal(id="main-container"):
            yield ServerTree(id="server-tree")
            yield DetailPanel(id="detail-panel")
        yield Footer()

    def on_mount(self) -> None:
        config = load_config()
        if config.ui.vim_keys:
            self.query_one(ServerTree).enable_vim_keys()
        self._show_server_picker()

    def _show_server_picker(self) -> None:
        self.app.push_screen(ServerPicker(), callback=self._on_server_selected)

    def _on_server_selected(self, server: Server | None) -> None:
        if server is not None:
            self._current_server = server
            self._load_server_with_sites(server)

    @work(exclusive=True, group="server-load")
    async def _load_server_with_sites(self, server: Server) -> None:
        tree = self.query_one(ServerTree)
        tree.loading = True
        try:
            api = SitesAPI(self.app.forge_client)
            sites = await api.list(server.id)
            tree.populate_server(server, sites)
        except ForgeAPIError as e:
            self.notify(f"Error loading server: {e}", severity="error", markup=False)
        except Exception as e:
            self.notify(f"Unexpected error: {e}", severity="error", markup=False)
        finally:
            tree.loading = False

    async def on_tree_node_selected(self, event: ServerTree.NodeSelected) -> None:
        node_data: NodeData | None = event.node.data
        if node_data is None:
            return

        # Skip group/root nodes that don't have panels
        if node_data.node_type in (NodeType.SERVER_ROOT, NodeType.SITE_ROOT, NodeType.SITES_GROUP):
            return

        detail = self.query_one(DetailPanel)
        await detail.show_panel(node_data)

    async def on_tree_node_expanded(self, event: ServerTree.NodeExpanded) -> None:
        node_data: NodeData | None = event.node.data
        if node_data is None:
            return

        if node_data.node_type == NodeType.SITES_GROUP and not node_data.loaded:
            self._load_sites(event.node, node_data.server_id)

    @work(exclusive=True, group="sites")
    async def _load_sites(self, sites_node, server_id: int) -> None:
        try:
            api = SitesAPI(self.app.forge_client)
            sites = await api.list(server_id)
            tree = self.query_one(ServerTree)
            tree.add_sites_to_node(sites_node, sites)
        except ForgeAPIError as e:
            self.notify(f"Error loading sites: {e}", severity="error", markup=False)

    def action_refresh(self) -> None:
        if self._current_server is not None:
            self._load_server_with_sites(self._current_server)

    def action_switch_server(self) -> None:
        self._show_server_picker()
