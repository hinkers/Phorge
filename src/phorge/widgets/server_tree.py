"""Left-panel tree widget for navigating Forge servers and resources."""

from __future__ import annotations

from dataclasses import dataclass
from enum import Enum, auto

from textual.binding import Binding
from textual.widgets import Tree
from textual.widgets._tree import TreeNode

from phorge.api.models import Server, Site


class NodeType(Enum):
    """Identifies what kind of data a tree node represents."""

    SERVER_ROOT = auto()
    SERVER_INFO = auto()
    SITES_GROUP = auto()
    SITE_ROOT = auto()
    SITE_INFO = auto()
    DEPLOYMENTS = auto()
    DEPLOYMENT_SCRIPT = auto()
    LOGS = auto()
    ENVIRONMENT = auto()
    WORKERS = auto()
    BACKUPS = auto()
    DOMAINS = auto()
    DATABASES_SITE = auto()
    SSL_CERTIFICATES = auto()
    COMMANDS = auto()
    GIT_REPOSITORY = auto()
    SSH_KEYS = auto()
    DAEMONS = auto()
    FIREWALL_RULES = auto()
    SCHEDULED_JOBS = auto()
    DATABASES_SERVER = auto()
    DATABASE_USERS = auto()


@dataclass
class NodeData:
    """Data attached to each tree node."""

    node_type: NodeType
    server_id: int
    site_id: int | None = None
    label: str = ""
    server_ip: str | None = None
    ssh_port: int = 22
    site_name: str | None = None
    site_directory: str | None = None
    loaded: bool = False


class ServerTree(Tree[NodeData]):
    """Navigation tree for Forge servers and their resources."""

    DEFAULT_CSS = """
    ServerTree {
        width: 1fr;
        min-width: 30;
        max-width: 55;
        border-right: solid $primary;
        scrollbar-gutter: stable;
    }
    """

    def __init__(self, **kwargs) -> None:
        super().__init__("Servers", **kwargs)
        self.show_root = True
        self.guide_depth = 3

    def enable_vim_keys(self) -> None:
        """Add vim-style keybindings to the tree."""
        self._bindings.bind("j", "cursor_down", "Down", show=False)
        self._bindings.bind("k", "cursor_up", "Up", show=False)
        self._bindings.bind("l", "select_cursor", "Expand/Select", show=False)
        self._bindings.bind("h", "cursor_parent", "Collapse/Parent", show=False)
        self._bindings.bind("g", "scroll_home", "Top", show=False)
        self._bindings.bind("G", "scroll_end", "Bottom", show=False)
        self._bindings.bind("o", "toggle_node", "Toggle", show=False)

    def action_cursor_parent(self) -> None:
        """Move to parent node or collapse current node."""
        node = self.cursor_node
        if node is None:
            return
        if node.is_expanded:
            node.collapse()
        elif node.parent is not None:
            self.select_node(node.parent)
            self.scroll_to_node(node.parent)

    def populate_servers(self, servers: list[Server]) -> None:
        """Clear tree and rebuild from server list."""
        self.clear()
        for server in servers:
            self._add_server_node(server)
        self.root.expand()

    def populate_server(self, server: Server, sites: list[Site]) -> None:
        """Clear tree and show a single server with its sites preloaded."""
        self.clear()
        ip_display = server.ip_address or "no ip"
        self.root.set_label(f"[bold]{server.name}[/bold] ({ip_display})")
        self.root.data = NodeData(
            NodeType.SERVER_ROOT,
            server.id,
            label=server.name,
            server_ip=server.ip_address,
            ssh_port=server.ssh_port,
        )

        self.root.add_leaf(
            "ℹ Server Info",
            data=NodeData(
                NodeType.SERVER_INFO,
                server.id,
                label=server.name,
                server_ip=server.ip_address,
                ssh_port=server.ssh_port,
            ),
        )

        sites_node = self.root.add(
            "Sites",
            data=NodeData(
                NodeType.SITES_GROUP,
                server.id,
                label="Sites",
                server_ip=server.ip_address,
                ssh_port=server.ssh_port,
                loaded=True,
            ),
        )
        self.add_sites_to_node(sites_node, sites)

        self.root.add_leaf(
            "SSH Keys",
            data=NodeData(NodeType.SSH_KEYS, server.id, label="SSH Keys"),
        )
        self.root.add_leaf(
            "Daemons",
            data=NodeData(NodeType.DAEMONS, server.id, label="Daemons"),
        )
        self.root.add_leaf(
            "Firewall Rules",
            data=NodeData(NodeType.FIREWALL_RULES, server.id, label="Firewall Rules"),
        )
        self.root.add_leaf(
            "Scheduled Jobs",
            data=NodeData(NodeType.SCHEDULED_JOBS, server.id, label="Scheduled Jobs"),
        )
        self.root.add_leaf(
            "Databases",
            data=NodeData(NodeType.DATABASES_SERVER, server.id, label="Databases"),
        )
        self.root.add_leaf(
            "Database Users",
            data=NodeData(NodeType.DATABASE_USERS, server.id, label="Database Users"),
        )

        self.root.expand()
        sites_node.expand()

    def _add_server_node(self, server: Server) -> None:
        ip_display = server.ip_address or "no ip"
        server_node = self.root.add(
            f"[bold]{server.name}[/bold] ({ip_display})",
            data=NodeData(
                NodeType.SERVER_ROOT,
                server.id,
                label=server.name,
                server_ip=server.ip_address,
                ssh_port=server.ssh_port,
            ),
            expand=False,
        )

        server_node.add_leaf(
            "ℹ Server Info",
            data=NodeData(
                NodeType.SERVER_INFO,
                server.id,
                label=server.name,
                server_ip=server.ip_address,
                ssh_port=server.ssh_port,
            ),
        )

        server_node.add(
            "Sites",
            data=NodeData(
                NodeType.SITES_GROUP,
                server.id,
                label="Sites",
                server_ip=server.ip_address,
                ssh_port=server.ssh_port,
            ),
        )

        server_node.add_leaf(
            "SSH Keys",
            data=NodeData(NodeType.SSH_KEYS, server.id, label="SSH Keys"),
        )
        server_node.add_leaf(
            "Daemons",
            data=NodeData(NodeType.DAEMONS, server.id, label="Daemons"),
        )
        server_node.add_leaf(
            "Firewall Rules",
            data=NodeData(NodeType.FIREWALL_RULES, server.id, label="Firewall Rules"),
        )
        server_node.add_leaf(
            "Scheduled Jobs",
            data=NodeData(NodeType.SCHEDULED_JOBS, server.id, label="Scheduled Jobs"),
        )
        server_node.add_leaf(
            "Databases",
            data=NodeData(NodeType.DATABASES_SERVER, server.id, label="Databases"),
        )
        server_node.add_leaf(
            "Database Users",
            data=NodeData(NodeType.DATABASE_USERS, server.id, label="Database Users"),
        )

    def add_sites_to_node(
        self, sites_node: TreeNode[NodeData], sites: list[Site]
    ) -> None:
        """Add site nodes under a Sites group node."""
        for site in sites:
            self._add_site_node(sites_node, site)
        if sites_node.data is not None:
            sites_node.data.loaded = True

    @staticmethod
    def _derive_site_directory(site: Site) -> str:
        """Derive the project root from web_directory, falling back to ~/site_name."""
        if site.web_directory and site.directory:
            # Strip the web subdirectory suffix (e.g. /public) to get project root
            suffix = site.directory.rstrip("/")
            if site.web_directory.endswith(suffix):
                return site.web_directory[: -len(suffix)].rstrip("/")
        return f"/home/forge/{site.name}"

    def _add_site_node(
        self, sites_node: TreeNode[NodeData], site: Site
    ) -> None:
        server_id = site.server_id
        site_id = site.id
        # Inherit server_ip/ssh_port from the parent Sites group node
        parent_data = sites_node.data
        server_ip = parent_data.server_ip if parent_data else None
        ssh_port = parent_data.ssh_port if parent_data else 22
        site_dir = self._derive_site_directory(site)

        site_node = sites_node.add(
            site.name,
            data=NodeData(
                NodeType.SITE_ROOT, server_id, site_id, site.name,
                server_ip=server_ip, ssh_port=ssh_port, site_name=site.name,
                site_directory=site_dir,
            ),
        )

        sub_items = [
            ("ℹ Site Info", NodeType.SITE_INFO),
            ("Deployments", NodeType.DEPLOYMENTS),
            ("Deployment Script", NodeType.DEPLOYMENT_SCRIPT),
            ("Logs", NodeType.LOGS),
            ("Environment File", NodeType.ENVIRONMENT),
            ("Workers", NodeType.WORKERS),
            ("Backups", NodeType.BACKUPS),
            ("Domains", NodeType.DOMAINS),
            ("Databases", NodeType.DATABASES_SITE),
            ("SSL Certificates", NodeType.SSL_CERTIFICATES),
            ("Commands", NodeType.COMMANDS),
            ("Git Repository", NodeType.GIT_REPOSITORY),
        ]
        for label, node_type in sub_items:
            site_node.add_leaf(
                label,
                data=NodeData(
                    node_type, server_id, site_id, label,
                    server_ip=server_ip, ssh_port=ssh_port, site_name=site.name,
                    site_directory=site_dir,
                ),
            )
