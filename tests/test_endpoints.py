"""Tests for API endpoint modules."""

from __future__ import annotations

from unittest.mock import AsyncMock

import pytest

from phorge.api.endpoints.servers import ServersAPI
from phorge.api.endpoints.sites import SitesAPI
from phorge.api.endpoints.deployments import DeploymentsAPI
from phorge.api.endpoints.databases import DatabasesAPI
from phorge.api.endpoints.ssh_keys import SSHKeysAPI
from phorge.api.endpoints.workers import WorkersAPI
from phorge.api.endpoints.daemons import DaemonsAPI
from phorge.api.endpoints.firewall import FirewallAPI
from phorge.api.endpoints.jobs import ScheduledJobsAPI
from phorge.api.endpoints.commands import CommandsAPI
from phorge.api.endpoints.backups import BackupsAPI
from phorge.api.endpoints.environment import EnvironmentAPI
from phorge.api.endpoints.logs import LogsAPI
from phorge.api.endpoints.ssl_certificates import SSLCertificatesAPI


class TestServersAPI:
    @pytest.mark.asyncio
    async def test_list(self, mock_client, sample_server_data):
        mock_client.get.return_value = {"servers": [sample_server_data]}
        api = ServersAPI(mock_client)
        servers = await api.list()
        assert len(servers) == 1
        assert servers[0].name == "production"
        mock_client.get.assert_called_once_with("/servers")

    @pytest.mark.asyncio
    async def test_list_empty(self, mock_client):
        mock_client.get.return_value = {"servers": []}
        api = ServersAPI(mock_client)
        servers = await api.list()
        assert servers == []

    @pytest.mark.asyncio
    async def test_get(self, mock_client, sample_server_data):
        mock_client.get.return_value = {"server": sample_server_data}
        api = ServersAPI(mock_client)
        server = await api.get(1)
        assert server.id == 1
        assert server.name == "production"

    @pytest.mark.asyncio
    async def test_reboot(self, mock_client):
        api = ServersAPI(mock_client)
        await api.reboot(1)
        mock_client.post.assert_called_once_with("/servers/1/reboot")

    @pytest.mark.asyncio
    async def test_get_user(self, mock_client):
        mock_client.get.return_value = {"user": {"id": 1, "name": "Test", "email": "test@test.com"}}
        api = ServersAPI(mock_client)
        user = await api.get_user()
        assert user.name == "Test"


class TestSitesAPI:
    @pytest.mark.asyncio
    async def test_list(self, mock_client, sample_site_data):
        mock_client.get.return_value = {"sites": [sample_site_data]}
        api = SitesAPI(mock_client)
        sites = await api.list(1)
        assert len(sites) == 1
        assert sites[0].name == "example.com"

    @pytest.mark.asyncio
    async def test_get(self, mock_client, sample_site_data):
        mock_client.get.return_value = {"site": sample_site_data}
        api = SitesAPI(mock_client)
        site = await api.get(1, 10)
        assert site.id == 10

    @pytest.mark.asyncio
    async def test_update_aliases(self, mock_client, sample_site_data):
        mock_client.put.return_value = {"site": sample_site_data}
        api = SitesAPI(mock_client)
        site = await api.update_aliases(1, 10, ["alias.com"])
        mock_client.put.assert_called_once_with(
            "/servers/1/sites/10/aliases",
            json={"aliases": ["alias.com"]},
        )


class TestDeploymentsAPI:
    @pytest.mark.asyncio
    async def test_list(self, mock_client):
        mock_client.get.return_value = {
            "deployments": [
                {"id": 1, "server_id": 1, "site_id": 10, "status": "finished"}
            ]
        }
        api = DeploymentsAPI(mock_client)
        deps = await api.list(1, 10)
        assert len(deps) == 1
        assert deps[0].status == "finished"

    @pytest.mark.asyncio
    async def test_get_output(self, mock_client):
        mock_client.get.return_value = {"output": "Deploying...done"}
        api = DeploymentsAPI(mock_client)
        output = await api.get_output(1, 10, 100)
        assert output == "Deploying...done"

    @pytest.mark.asyncio
    async def test_get_script(self, mock_client):
        mock_client.get_text.return_value = "#!/bin/bash\ncd /home/forge"
        api = DeploymentsAPI(mock_client)
        script = await api.get_script(1, 10)
        assert script == "#!/bin/bash\ncd /home/forge"
        mock_client.get_text.assert_called_once()

    @pytest.mark.asyncio
    async def test_deploy(self, mock_client):
        api = DeploymentsAPI(mock_client)
        await api.deploy(1, 10)
        mock_client.post.assert_called_once_with("/servers/1/sites/10/deployment/deploy")

    @pytest.mark.asyncio
    async def test_reset_status(self, mock_client):
        api = DeploymentsAPI(mock_client)
        await api.reset_status(1, 10)
        mock_client.post.assert_called_once_with("/servers/1/sites/10/deployment/reset")

    @pytest.mark.asyncio
    async def test_update_script(self, mock_client):
        api = DeploymentsAPI(mock_client)
        await api.update_script(1, 10, "new script")
        mock_client.put.assert_called_once_with(
            "/servers/1/sites/10/deployment/script",
            json={"content": "new script"},
        )


class TestEnvironmentAPI:
    @pytest.mark.asyncio
    async def test_get(self, mock_client):
        mock_client.get_text.return_value = "APP_NAME=Laravel\nAPP_ENV=production"
        api = EnvironmentAPI(mock_client)
        env = await api.get(1, 10)
        assert "APP_NAME=Laravel" in env
        mock_client.get_text.assert_called_once()

    @pytest.mark.asyncio
    async def test_update(self, mock_client):
        api = EnvironmentAPI(mock_client)
        await api.update(1, 10, "APP_NAME=Test")
        mock_client.put.assert_called_once_with(
            "/servers/1/sites/10/env",
            json={"content": "APP_NAME=Test"},
        )


class TestDatabasesAPI:
    @pytest.mark.asyncio
    async def test_list(self, mock_client):
        mock_client.get.return_value = {
            "databases": [{"id": 1, "server_id": 1, "name": "forge"}]
        }
        api = DatabasesAPI(mock_client)
        dbs = await api.list(1)
        assert len(dbs) == 1
        assert dbs[0].name == "forge"

    @pytest.mark.asyncio
    async def test_create(self, mock_client):
        mock_client.post.return_value = {"database": {"id": 2, "server_id": 1, "name": "new_db"}}
        api = DatabasesAPI(mock_client)
        await api.create(1, "new_db")
        mock_client.post.assert_called_once_with(
            "/servers/1/databases",
            json={"name": "new_db"},
        )

    @pytest.mark.asyncio
    async def test_delete(self, mock_client):
        api = DatabasesAPI(mock_client)
        await api.delete(1, 5)
        mock_client.delete.assert_called_once_with("/servers/1/databases/5")


class TestSSHKeysAPI:
    @pytest.mark.asyncio
    async def test_list(self, mock_client):
        mock_client.get.return_value = {
            "keys": [{"id": 1, "name": "my-key"}]
        }
        api = SSHKeysAPI(mock_client)
        keys = await api.list(1)
        assert len(keys) == 1
        assert keys[0].name == "my-key"

    @pytest.mark.asyncio
    async def test_create(self, mock_client):
        mock_client.post.return_value = {"key": {"id": 2, "name": "new-key"}}
        api = SSHKeysAPI(mock_client)
        await api.create(1, "new-key", "ssh-ed25519 AAAA...")
        mock_client.post.assert_called_once_with(
            "/servers/1/keys",
            json={"name": "new-key", "key": "ssh-ed25519 AAAA...", "username": "forge"},
        )

    @pytest.mark.asyncio
    async def test_delete(self, mock_client):
        api = SSHKeysAPI(mock_client)
        await api.delete(1, 5)
        mock_client.delete.assert_called_once_with("/servers/1/keys/5")


class TestLogsAPI:
    @pytest.mark.asyncio
    async def test_get_server_logs(self, mock_client):
        mock_client.get.return_value = {"content": "log line 1\nlog line 2"}
        api = LogsAPI(mock_client)
        logs = await api.get_server_logs(1)
        assert "log line 1" in logs

    @pytest.mark.asyncio
    async def test_get_site_logs(self, mock_client):
        mock_client.get.return_value = {"content": "site log"}
        api = LogsAPI(mock_client)
        logs = await api.get_site_logs(1, 10)
        assert logs == "site log"


class TestSSLCertificatesAPI:
    @pytest.mark.asyncio
    async def test_list(self, mock_client):
        mock_client.get.return_value = {
            "certificates": [{"id": 1, "domain": "example.com", "active": True}]
        }
        api = SSLCertificatesAPI(mock_client)
        certs = await api.list(1, 10)
        assert len(certs) == 1
        assert certs[0].domain == "example.com"

    @pytest.mark.asyncio
    async def test_obtain_letsencrypt(self, mock_client):
        mock_client.post.return_value = {"certificate": {"id": 2}}
        api = SSLCertificatesAPI(mock_client)
        await api.obtain_letsencrypt(1, 10, ["example.com"])
        mock_client.post.assert_called_once_with(
            "/servers/1/sites/10/certificates/letsencrypt",
            json={"domains": ["example.com"]},
        )


class TestFirewallAPI:
    @pytest.mark.asyncio
    async def test_list(self, mock_client):
        mock_client.get.return_value = {
            "rules": [{"id": 1, "server_id": 1, "name": "SSH", "port": 22}]
        }
        api = FirewallAPI(mock_client)
        rules = await api.list(1)
        assert len(rules) == 1
        assert rules[0].name == "SSH"

    @pytest.mark.asyncio
    async def test_delete(self, mock_client):
        api = FirewallAPI(mock_client)
        await api.delete(1, 5)
        mock_client.delete.assert_called_once_with("/servers/1/firewall-rules/5")


class TestDaemonsAPI:
    @pytest.mark.asyncio
    async def test_list(self, mock_client):
        mock_client.get.return_value = {
            "daemons": [{"id": 1, "server_id": 1, "command": "php artisan queue:work"}]
        }
        api = DaemonsAPI(mock_client)
        daemons = await api.list(1)
        assert len(daemons) == 1
        assert "queue:work" in daemons[0].command


class TestScheduledJobsAPI:
    @pytest.mark.asyncio
    async def test_list(self, mock_client):
        mock_client.get.return_value = {
            "jobs": [{"id": 1, "server_id": 1, "command": "php artisan schedule:run"}]
        }
        api = ScheduledJobsAPI(mock_client)
        jobs = await api.list(1)
        assert len(jobs) == 1


class TestWorkersAPI:
    @pytest.mark.asyncio
    async def test_list(self, mock_client):
        mock_client.get.return_value = {
            "workers": [{"id": 1, "connection": "redis", "queue": "default"}]
        }
        api = WorkersAPI(mock_client)
        workers = await api.list(1, 10)
        assert len(workers) == 1
        assert workers[0].connection == "redis"

    @pytest.mark.asyncio
    async def test_restart(self, mock_client):
        api = WorkersAPI(mock_client)
        await api.restart(1, 10, 5)
        mock_client.post.assert_called_once_with("/servers/1/sites/10/workers/5/restart")


class TestCommandsAPI:
    @pytest.mark.asyncio
    async def test_execute(self, mock_client):
        mock_client.post.return_value = {"command": {"id": 1, "server_id": 1, "site_id": 10, "command": "ls"}}
        api = CommandsAPI(mock_client)
        await api.execute(1, 10, "ls")
        mock_client.post.assert_called_once_with(
            "/servers/1/sites/10/commands",
            json={"command": "ls"},
        )


class TestBackupsAPI:
    @pytest.mark.asyncio
    async def test_list_configs(self, mock_client):
        mock_client.get.return_value = {
            "backups": [{"id": 1, "server_id": 1, "provider": "s3"}]
        }
        api = BackupsAPI(mock_client)
        configs = await api.list_configs(1)
        assert len(configs) == 1
        assert configs[0].provider == "s3"
