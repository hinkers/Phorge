"""Right-side detail panel that swaps content based on tree selection."""

from __future__ import annotations

from textual.containers import VerticalScroll
from textual.widget import Widget
from textual.widgets import Static

from phorge.widgets.server_tree import NodeData, NodeType


class DetailPanel(VerticalScroll):
    """Right-side panel that displays content based on tree selection."""

    DEFAULT_CSS = """
    DetailPanel {
        width: 3fr;
        padding: 1 2;
    }
    """

    def __init__(self, **kwargs) -> None:
        super().__init__(**kwargs)
        self._current_panel: Widget | None = None

    def compose(self):
        yield Static(
            "[dim]Select an item from the tree to view details.[/dim]",
            id="placeholder",
        )

    async def show_panel(self, node_data: NodeData) -> None:
        """Swap displayed panel based on selected node data."""
        # Remove existing content
        if self._current_panel is not None:
            await self._current_panel.remove()
            self._current_panel = None

        placeholder = self.query("#placeholder")
        for widget in placeholder:
            await widget.remove()

        panel_class = _get_panel_class(node_data.node_type)
        if panel_class is None:
            fallback = Static(f"[dim]No panel for {node_data.node_type.name}[/dim]")
            await self.mount(fallback)
            self._current_panel = fallback
            return

        panel = panel_class(node_data=node_data)
        await self.mount(panel)
        self._current_panel = panel


def _get_panel_class(node_type: NodeType) -> type[Widget] | None:
    """Lazy import and return the panel class for a given node type."""
    from phorge.widgets.panels.server_info import ServerInfoPanel
    from phorge.widgets.panels.site_info import SiteInfoPanel
    from phorge.widgets.panels.deployments import DeploymentsPanel
    from phorge.widgets.panels.deployment_script import DeploymentScriptPanel
    from phorge.widgets.panels.logs_panel import LogsPanel
    from phorge.widgets.panels.env_panel import EnvironmentPanel
    from phorge.widgets.panels.workers_panel import WorkersPanel
    from phorge.widgets.panels.backups_panel import BackupsPanel
    from phorge.widgets.panels.domains_panel import DomainsPanel
    from phorge.widgets.panels.databases_panel import DatabasesPanel
    from phorge.widgets.panels.db_users_panel import DatabaseUsersPanel
    from phorge.widgets.panels.ssl_panel import SSLPanel
    from phorge.widgets.panels.commands_panel import CommandsPanel
    from phorge.widgets.panels.git_panel import GitPanel
    from phorge.widgets.panels.ssh_keys_panel import SSHKeysPanel
    from phorge.widgets.panels.daemons_panel import DaemonsPanel
    from phorge.widgets.panels.firewall_panel import FirewallPanel
    from phorge.widgets.panels.jobs_panel import ScheduledJobsPanel

    registry: dict[NodeType, type[Widget]] = {
        NodeType.SERVER_INFO: ServerInfoPanel,
        NodeType.SITE_INFO: SiteInfoPanel,
        NodeType.DEPLOYMENTS: DeploymentsPanel,
        NodeType.DEPLOYMENT_SCRIPT: DeploymentScriptPanel,
        NodeType.LOGS: LogsPanel,
        NodeType.ENVIRONMENT: EnvironmentPanel,
        NodeType.WORKERS: WorkersPanel,
        NodeType.BACKUPS: BackupsPanel,
        NodeType.DOMAINS: DomainsPanel,
        NodeType.DATABASES_SERVER: DatabasesPanel,
        NodeType.DATABASES_SITE: DatabasesPanel,
        NodeType.DATABASE_USERS: DatabaseUsersPanel,
        NodeType.SSL_CERTIFICATES: SSLPanel,
        NodeType.COMMANDS: CommandsPanel,
        NodeType.GIT_REPOSITORY: GitPanel,
        NodeType.SSH_KEYS: SSHKeysPanel,
        NodeType.DAEMONS: DaemonsPanel,
        NodeType.FIREWALL_RULES: FirewallPanel,
        NodeType.SCHEDULED_JOBS: ScheduledJobsPanel,
    }

    return registry.get(node_type)
