"""Textual pilot tests for detail panels.

Each panel is mounted inside a minimal App with a mocked ForgeClient,
verifying that compose works, load_data fetches and displays content,
and buttons/actions exist.
"""

from __future__ import annotations

from unittest.mock import AsyncMock, patch

import pytest

from textual.app import App, ComposeResult
from textual.widgets import Static, DataTable, Button

from phorge.api.client import ForgeClient
from phorge.widgets.server_tree import NodeData, NodeType


# ---------------------------------------------------------------------------
# Helper: minimal app that mounts a single panel with a mocked ForgeClient
# ---------------------------------------------------------------------------

def _make_mock_client() -> ForgeClient:
    client = ForgeClient("test-key")
    client.get = AsyncMock(return_value={})
    client.get_text = AsyncMock(return_value="")
    client.post = AsyncMock(return_value={})
    client.put = AsyncMock(return_value={})
    client.delete = AsyncMock(return_value={})
    return client


class PanelTestApp(App):
    """Minimal app host for testing a single panel widget."""

    CSS = """
    Screen { layout: vertical; }
    """

    def __init__(self, panel_widget, **kwargs):
        super().__init__(**kwargs)
        self._panel = panel_widget
        self.forge_client = _make_mock_client()

    def compose(self) -> ComposeResult:
        yield self._panel


# ---------------------------------------------------------------------------
# Fixtures
# ---------------------------------------------------------------------------

@pytest.fixture
def server_nd():
    return NodeData(NodeType.SERVER_INFO, server_id=1, label="prod", server_ip="1.2.3.4", ssh_port=22)


@pytest.fixture
def site_nd():
    return NodeData(NodeType.DEPLOYMENTS, server_id=1, site_id=10, label="example.com")


# ---------------------------------------------------------------------------
# ServerInfoPanel
# ---------------------------------------------------------------------------

class TestServerInfoPanel:
    @pytest.mark.asyncio
    async def test_composes_and_loads(self, server_nd):
        from phorge.widgets.panels.server_info import ServerInfoPanel
        from phorge.api.models import Server

        panel = ServerInfoPanel(node_data=server_nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "server": {
                "id": 1, "name": "prod", "ip_address": "1.2.3.4", "ssh_port": 22,
                "is_ready": True, "status": "installed",
            }
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            # Verify title is rendered
            title = app.query_one(".panel-title", Static)
            assert "prod" in title.content

            # Verify buttons exist
            ssh_btn = app.query_one("#btn-ssh", Button)
            assert ssh_btn is not None
            reboot_btn = app.query_one("#btn-reboot", Button)
            assert reboot_btn is not None

            # load_data should have called the API
            app.forge_client.get.assert_called()

    @pytest.mark.asyncio
    async def test_displays_server_data(self, server_nd):
        from phorge.widgets.panels.server_info import ServerInfoPanel

        panel = ServerInfoPanel(node_data=server_nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "server": {
                "id": 1, "name": "prod", "ip_address": "1.2.3.4", "ssh_port": 22,
                "is_ready": True, "status": "installed", "provider": "ocean2",
                "type": "app", "php_version": "8.2",
            }
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            content = app.query_one("#server-info-content", Static)
            rendered = content.content
            assert "1.2.3.4" in rendered
            assert "prod" in rendered

    @pytest.mark.asyncio
    async def test_handles_api_error(self, server_nd):
        from phorge.widgets.panels.server_info import ServerInfoPanel
        from phorge.api.exceptions import ForgeAPIError

        panel = ServerInfoPanel(node_data=server_nd)
        app = PanelTestApp(panel)
        app.forge_client.get.side_effect = ForgeAPIError("Connection failed")

        async with app.run_test() as pilot:
            await pilot.pause()
            content = app.query_one("#server-info-content", Static)
            rendered = content.content
            assert "Error" in rendered


# ---------------------------------------------------------------------------
# SiteInfoPanel
# ---------------------------------------------------------------------------

class TestSiteInfoPanel:
    @pytest.mark.asyncio
    async def test_composes_and_loads(self):
        from phorge.widgets.panels.site_info import SiteInfoPanel

        nd = NodeData(NodeType.SITE_INFO, server_id=1, site_id=10, label="example.com")
        panel = SiteInfoPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "site": {
                "id": 10, "server_id": 1, "name": "example.com",
                "status": "installed", "directory": "/home/forge/example.com",
            }
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            title = app.query_one(".panel-title", Static)
            assert "example.com" in title.content
            app.forge_client.get.assert_called()


# ---------------------------------------------------------------------------
# DeploymentsPanel
# ---------------------------------------------------------------------------

class TestDeploymentsPanel:
    @pytest.mark.asyncio
    async def test_composes_with_table(self, site_nd):
        from phorge.widgets.panels.deployments import DeploymentsPanel

        panel = DeploymentsPanel(node_data=site_nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "deployments": [
                {"id": 1, "server_id": 1, "site_id": 10, "status": "finished",
                 "commit_hash": "abc123", "commit_author": "dev",
                 "commit_message": "fix bug", "started_at": "2024-01-01T00:00:00Z"},
            ]
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            table = app.query_one("#deployments-table", DataTable)
            assert table.row_count == 1

            # Buttons should exist
            deploy_btn = app.query_one("#btn-deploy", Button)
            assert deploy_btn is not None
            reset_btn = app.query_one("#btn-reset", Button)
            assert reset_btn is not None

    @pytest.mark.asyncio
    async def test_empty_deployments(self, site_nd):
        from phorge.widgets.panels.deployments import DeploymentsPanel

        panel = DeploymentsPanel(node_data=site_nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {"deployments": []}

        async with app.run_test() as pilot:
            await pilot.pause()
            table = app.query_one("#deployments-table", DataTable)
            assert table.row_count == 0


# ---------------------------------------------------------------------------
# DeploymentScriptPanel
# ---------------------------------------------------------------------------

class TestDeploymentScriptPanel:
    @pytest.mark.asyncio
    async def test_composes_and_loads_script(self):
        from phorge.widgets.panels.deployment_script import DeploymentScriptPanel

        nd = NodeData(NodeType.DEPLOYMENT_SCRIPT, server_id=1, site_id=10, label="Script")
        panel = DeploymentScriptPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get_text.return_value = "#!/bin/bash\ncd /home/forge\ngit pull"

        async with app.run_test() as pilot:
            await pilot.pause()
            content = app.query_one("#script-content", Static)
            rendered = content.content
            assert "git pull" in rendered
            app.forge_client.get_text.assert_called()


# ---------------------------------------------------------------------------
# LogsPanel
# ---------------------------------------------------------------------------

class TestLogsPanel:
    @pytest.mark.asyncio
    async def test_loads_site_logs(self):
        from phorge.widgets.panels.logs_panel import LogsPanel

        nd = NodeData(NodeType.LOGS, server_id=1, site_id=10, label="Logs")
        panel = LogsPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {"content": "Error at line 42\nStack trace here"}

        async with app.run_test() as pilot:
            await pilot.pause()
            content = app.query_one("#log-content", Static)
            rendered = content.content
            assert "Error at line 42" in rendered

    @pytest.mark.asyncio
    async def test_loads_server_logs(self):
        from phorge.widgets.panels.logs_panel import LogsPanel

        nd = NodeData(NodeType.LOGS, server_id=1, label="Logs")  # no site_id → server logs
        panel = LogsPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {"content": "Server log line 1"}

        async with app.run_test() as pilot:
            await pilot.pause()
            content = app.query_one("#log-content", Static)
            rendered = content.content
            assert "Server log line 1" in rendered

    @pytest.mark.asyncio
    async def test_has_clear_button_for_site(self):
        from phorge.widgets.panels.logs_panel import LogsPanel

        nd = NodeData(NodeType.LOGS, server_id=1, site_id=10, label="Logs")
        panel = LogsPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {"content": "logs"}

        async with app.run_test() as pilot:
            await pilot.pause()
            clear_btn = app.query_one("#btn-clear", Button)
            assert clear_btn is not None


# ---------------------------------------------------------------------------
# EnvironmentPanel
# ---------------------------------------------------------------------------

class TestEnvironmentPanel:
    @pytest.mark.asyncio
    async def test_loads_env_content(self):
        from phorge.widgets.panels.env_panel import EnvironmentPanel

        nd = NodeData(NodeType.ENVIRONMENT, server_id=1, site_id=10, label="Environment")
        panel = EnvironmentPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get_text.return_value = "APP_NAME=Laravel\nAPP_ENV=production"

        async with app.run_test() as pilot:
            await pilot.pause()
            content = app.query_one("#env-content", Static)
            rendered = content.content
            assert "APP_NAME=Laravel" in rendered
            app.forge_client.get_text.assert_called()


# ---------------------------------------------------------------------------
# DatabasesPanel
# ---------------------------------------------------------------------------

class TestDatabasesPanel:
    @pytest.mark.asyncio
    async def test_loads_databases(self):
        from phorge.widgets.panels.databases_panel import DatabasesPanel

        nd = NodeData(NodeType.DATABASES_SERVER, server_id=1, label="Databases")
        panel = DatabasesPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "databases": [
                {"id": 1, "server_id": 1, "name": "forge", "status": "installed", "is_synced": True},
                {"id": 2, "server_id": 1, "name": "myapp", "status": "installed", "is_synced": False},
            ]
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            table = app.query_one("#databases-table", DataTable)
            assert table.row_count == 2

    @pytest.mark.asyncio
    async def test_has_create_button(self):
        from phorge.widgets.panels.databases_panel import DatabasesPanel

        nd = NodeData(NodeType.DATABASES_SERVER, server_id=1, label="Databases")
        panel = DatabasesPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {"databases": []}

        async with app.run_test() as pilot:
            await pilot.pause()
            create_btn = app.query_one("#btn-create", Button)
            assert create_btn is not None


# ---------------------------------------------------------------------------
# SSHKeysPanel
# ---------------------------------------------------------------------------

class TestSSHKeysPanel:
    @pytest.mark.asyncio
    async def test_loads_ssh_keys(self):
        from phorge.widgets.panels.ssh_keys_panel import SSHKeysPanel

        nd = NodeData(NodeType.SSH_KEYS, server_id=1, label="SSH Keys")
        panel = SSHKeysPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "keys": [
                {"id": 1, "name": "my-key", "status": "installed"},
                {"id": 2, "name": "deploy-key", "status": "installed"},
            ]
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            table = app.query_one("#ssh-keys-table", DataTable)
            assert table.row_count == 2

    @pytest.mark.asyncio
    async def test_has_add_buttons(self):
        from phorge.widgets.panels.ssh_keys_panel import SSHKeysPanel

        nd = NodeData(NodeType.SSH_KEYS, server_id=1, label="SSH Keys")
        panel = SSHKeysPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {"keys": []}

        async with app.run_test() as pilot:
            await pilot.pause()
            add_btn = app.query_one("#btn-add", Button)
            assert add_btn is not None
            local_btn = app.query_one("#btn-add-local", Button)
            assert local_btn is not None


# ---------------------------------------------------------------------------
# WorkersPanel
# ---------------------------------------------------------------------------

class TestWorkersPanel:
    @pytest.mark.asyncio
    async def test_loads_workers(self):
        from phorge.widgets.panels.workers_panel import WorkersPanel

        nd = NodeData(NodeType.WORKERS, server_id=1, site_id=10, label="Workers")
        panel = WorkersPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "workers": [
                {"id": 1, "connection": "redis", "queue": "default",
                 "processes": 3, "timeout": 60, "status": "running"},
            ]
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            table = app.query_one("#workers-table", DataTable)
            assert table.row_count == 1


# ---------------------------------------------------------------------------
# BackupsPanel
# ---------------------------------------------------------------------------

class TestBackupsPanel:
    @pytest.mark.asyncio
    async def test_loads_backups(self):
        from phorge.widgets.panels.backups_panel import BackupsPanel

        nd = NodeData(NodeType.BACKUPS, server_id=1, site_id=10, label="Backups")
        panel = BackupsPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "backups": [
                {"id": 1, "server_id": 1, "provider": "s3",
                 "frequency": "daily", "time": "00:00", "day": None},
            ]
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            table = app.query_one("#backups-table", DataTable)
            assert table.row_count == 1


# ---------------------------------------------------------------------------
# DomainsPanel
# ---------------------------------------------------------------------------

class TestDomainsPanel:
    @pytest.mark.asyncio
    async def test_loads_domains(self):
        from phorge.widgets.panels.domains_panel import DomainsPanel

        nd = NodeData(NodeType.DOMAINS, server_id=1, site_id=10, label="Domains")
        panel = DomainsPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "site": {
                "id": 10, "server_id": 1, "name": "example.com",
                "aliases": ["www.example.com", "api.example.com"],
            }
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            table = app.query_one("#domains-table", DataTable)
            # Should have rows for each alias
            assert table.row_count >= 1


# ---------------------------------------------------------------------------
# SSLPanel
# ---------------------------------------------------------------------------

class TestSSLPanel:
    @pytest.mark.asyncio
    async def test_loads_certificates(self):
        from phorge.widgets.panels.ssl_panel import SSLPanel

        nd = NodeData(NodeType.SSL_CERTIFICATES, server_id=1, site_id=10, label="SSL")
        panel = SSLPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "certificates": [
                {"id": 1, "domain": "example.com", "active": True,
                 "type": "letsencrypt", "status": "installed"},
            ]
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            table = app.query_one("#ssl-table", DataTable)
            assert table.row_count == 1


# ---------------------------------------------------------------------------
# CommandsPanel
# ---------------------------------------------------------------------------

class TestCommandsPanel:
    @pytest.mark.asyncio
    async def test_loads_commands(self):
        from phorge.widgets.panels.commands_panel import CommandsPanel

        nd = NodeData(NodeType.COMMANDS, server_id=1, site_id=10, label="Commands")
        panel = CommandsPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "commands": [
                {"id": 1, "server_id": 1, "site_id": 10,
                 "command": "php artisan migrate", "status": "finished",
                 "user": "forge", "created_at": "2024-01-01", "duration": 5},
            ]
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            table = app.query_one("#commands-table", DataTable)
            assert table.row_count == 1

    @pytest.mark.asyncio
    async def test_has_execute_button(self):
        from phorge.widgets.panels.commands_panel import CommandsPanel

        nd = NodeData(NodeType.COMMANDS, server_id=1, site_id=10, label="Commands")
        panel = CommandsPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {"commands": []}

        async with app.run_test() as pilot:
            await pilot.pause()
            exec_btn = app.query_one("#btn-execute", Button)
            assert exec_btn is not None


# ---------------------------------------------------------------------------
# GitPanel
# ---------------------------------------------------------------------------

class TestGitPanel:
    @pytest.mark.asyncio
    async def test_loads_git_info(self):
        from phorge.widgets.panels.git_panel import GitPanel

        nd = NodeData(NodeType.GIT_REPOSITORY, server_id=1, site_id=10, label="Git")
        panel = GitPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "site": {
                "id": 10, "server_id": 1, "name": "example.com",
                "repository": "user/repo", "repository_provider": "github",
                "repository_branch": "main", "repository_status": "installed",
            }
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            info = app.query_one("#git-info", Static)
            rendered = info.content
            assert "user/repo" in rendered


# ---------------------------------------------------------------------------
# DaemonsPanel
# ---------------------------------------------------------------------------

class TestDaemonsPanel:
    @pytest.mark.asyncio
    async def test_loads_daemons(self):
        from phorge.widgets.panels.daemons_panel import DaemonsPanel

        nd = NodeData(NodeType.DAEMONS, server_id=1, label="Daemons")
        panel = DaemonsPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "daemons": [
                {"id": 1, "server_id": 1, "command": "php artisan queue:work",
                 "user": "forge", "processes": 1, "status": "running"},
            ]
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            table = app.query_one("#daemons-table", DataTable)
            assert table.row_count == 1


# ---------------------------------------------------------------------------
# ScheduledJobsPanel
# ---------------------------------------------------------------------------

class TestScheduledJobsPanel:
    @pytest.mark.asyncio
    async def test_loads_jobs(self):
        from phorge.widgets.panels.jobs_panel import ScheduledJobsPanel

        nd = NodeData(NodeType.SCHEDULED_JOBS, server_id=1, label="Jobs")
        panel = ScheduledJobsPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "jobs": [
                {"id": 1, "server_id": 1, "command": "php artisan schedule:run",
                 "user": "forge", "frequency": "minutely", "cron": "* * * * *",
                 "status": "active"},
            ]
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            table = app.query_one("#jobs-table", DataTable)
            assert table.row_count == 1


# ---------------------------------------------------------------------------
# FirewallPanel
# ---------------------------------------------------------------------------

class TestFirewallPanel:
    @pytest.mark.asyncio
    async def test_loads_firewall_rules(self):
        from phorge.widgets.panels.firewall_panel import FirewallPanel

        nd = NodeData(NodeType.FIREWALL_RULES, server_id=1, label="Firewall")
        panel = FirewallPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "rules": [
                {"id": 1, "server_id": 1, "name": "SSH", "port": 22,
                 "ip_address": None, "type": "allow", "status": "installed"},
            ]
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            table = app.query_one("#firewall-table", DataTable)
            assert table.row_count == 1


# ---------------------------------------------------------------------------
# DatabaseUsersPanel
# ---------------------------------------------------------------------------

class TestDatabaseUsersPanel:
    @pytest.mark.asyncio
    async def test_loads_db_users(self):
        from phorge.widgets.panels.db_users_panel import DatabaseUsersPanel

        nd = NodeData(NodeType.DATABASE_USERS, server_id=1, label="DB Users")
        panel = DatabaseUsersPanel(node_data=nd)
        app = PanelTestApp(panel)
        app.forge_client.get.return_value = {
            "users": [
                {"id": 1, "server_id": 1, "name": "forge",
                 "status": "installed", "databases": []},
            ]
        }

        async with app.run_test() as pilot:
            await pilot.pause()
            table = app.query_one("#db-users-table", DataTable)
            assert table.row_count == 1
