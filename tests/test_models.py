"""Tests for Pydantic models with Forge API edge cases."""

from __future__ import annotations

import pytest

from phorge.api.models import (
    Backup,
    BackupConfig,
    Certificate,
    Daemon,
    Database,
    DatabaseUser,
    Deployment,
    FirewallRule,
    GitProject,
    RedirectRule,
    ScheduledJob,
    Server,
    Site,
    SiteCommand,
    SSHKey,
    User,
    Worker,
)


class TestNoneToListCoercion:
    """Forge API often returns null for list fields - models must handle it."""

    def test_server_network_none(self):
        s = Server(id=1, name="test", network=None)
        assert s.network == []

    def test_server_tags_none(self):
        s = Server(id=1, name="test", tags=None)
        assert s.tags == []

    def test_site_aliases_none(self):
        s = Site(id=1, server_id=1, name="test", aliases=None)
        assert s.aliases == []

    def test_site_tags_none(self):
        s = Site(id=1, server_id=1, name="test", tags=None)
        assert s.tags == []

    def test_database_user_databases_none(self):
        u = DatabaseUser(id=1, server_id=1, name="forge", databases=None)
        assert u.databases == []

    def test_backup_config_databases_none(self):
        bc = BackupConfig(id=1, server_id=1, databases=None)
        assert bc.databases == []

    def test_backup_config_no_server_id(self):
        bc = BackupConfig(id=1)
        assert bc.server_id is None

    def test_backup_config_databases_as_dicts(self):
        """Forge API returns full database objects, not just IDs."""
        bc = BackupConfig(id=1, databases=[
            {"id": 100, "name": "forge", "created_at": "2026-02-10"},
            {"id": 101, "name": "myapp", "created_at": "2026-02-10"},
        ])
        assert len(bc.databases) == 2


class TestNoneToBoolCoercion:
    """Forge API sometimes returns null for bool fields."""

    def test_server_is_ready_none(self):
        s = Server(id=1, name="test", is_ready=None)
        assert s.is_ready is False

    def test_site_quick_deploy_none(self):
        s = Site(id=1, server_id=1, name="test", quick_deploy=None)
        assert s.quick_deploy is False

    def test_site_wildcards_none(self):
        s = Site(id=1, server_id=1, name="test", wildcards=None)
        assert s.wildcards is False

    def test_site_is_secured_none(self):
        s = Site(id=1, server_id=1, name="test", is_secured=None)
        assert s.is_secured is False

    def test_worker_daemon_none(self):
        w = Worker(id=1, daemon=None)
        assert w.daemon is False

    def test_worker_force_none(self):
        w = Worker(id=1, force=None)
        assert w.force is False

    def test_certificate_active_none(self):
        c = Certificate(id=1, active=None)
        assert c.active is False

    def test_certificate_existing_none(self):
        c = Certificate(id=1, existing=None)
        assert c.existing is False


class TestExtraFieldsIgnored:
    """Forge API may return fields not in our models."""

    def test_server_ignores_extra(self):
        s = Server(id=1, name="test", unknown_field="ignored", another=123)
        assert s.id == 1
        assert not hasattr(s, "unknown_field")

    def test_site_ignores_extra(self):
        s = Site(id=1, server_id=1, name="test", extra_data={"nested": True})
        assert s.id == 1

    def test_deployment_ignores_extra(self):
        d = Deployment(id=1, server_id=1, site_id=1, some_future_field="val")
        assert d.id == 1


class TestServerModel:
    def test_minimal(self):
        s = Server(id=1, name="test")
        assert s.id == 1
        assert s.name == "test"
        assert s.ip_address is None
        assert s.ssh_port == 22
        assert s.is_ready is False

    def test_full(self, sample_server_data):
        s = Server(**sample_server_data)
        assert s.id == 1
        assert s.name == "production"
        assert s.ip_address == "1.2.3.4"
        assert s.is_ready is True
        assert s.ssh_port == 22


class TestSiteModel:
    def test_minimal(self):
        s = Site(id=1, server_id=1, name="test.com")
        assert s.aliases == []
        assert s.quick_deploy is False

    def test_full(self, sample_site_data):
        s = Site(**sample_site_data)
        assert s.name == "example.com"
        assert s.aliases == ["www.example.com"]
        assert s.quick_deploy is True
        assert s.is_secured is True


class TestDeploymentModel:
    def test_minimal(self):
        d = Deployment(id=1, server_id=1, site_id=1)
        assert d.status is None
        assert d.commit_hash is None

    def test_full(self):
        d = Deployment(
            id=100,
            server_id=1,
            site_id=10,
            status="finished",
            commit_hash="abc123",
            commit_author="dev",
            commit_message="deploy fix",
            started_at="2024-01-01T00:00:00Z",
        )
        assert d.status == "finished"
        assert d.commit_hash == "abc123"


class TestDatabaseModel:
    def test_defaults(self):
        db = Database(id=1, server_id=1, name="forge")
        assert db.is_synced is True
        assert db.status is None


class TestSSHKeyModel:
    def test_minimal(self):
        k = SSHKey(id=1, name="my-key")
        assert k.status is None


class TestDaemonModel:
    def test_defaults(self):
        d = Daemon(id=1, server_id=1, command="php artisan queue:work")
        assert d.processes == 1
        assert d.user is None


class TestFirewallRuleModel:
    def test_string_port(self):
        r = FirewallRule(id=1, server_id=1, name="SSH", port="22")
        assert r.port == "22"

    def test_int_port(self):
        r = FirewallRule(id=1, server_id=1, name="SSH", port=22)
        assert r.port == 22


class TestWorkerModel:
    def test_defaults(self):
        w = Worker(id=1)
        assert w.timeout == 60
        assert w.sleep == 3
        assert w.processes == 1
        assert w.daemon is True

    def test_timeout_none(self):
        w = Worker(id=1, timeout=None)
        assert w.timeout == 60

    def test_sleep_none(self):
        w = Worker(id=1, sleep=None)
        assert w.sleep == 3

    def test_processes_none(self):
        w = Worker(id=1, processes=None)
        assert w.processes == 1


class TestRedirectRuleModel:
    def test_alias_field(self):
        r = RedirectRule(id=1, **{"from": "/old", "to": "/new", "type": "permanent"})
        assert r.from_url == "/old"
        assert r.to == "/new"


class TestSiteCommandModel:
    def test_minimal(self):
        c = SiteCommand(id=1, server_id=1, site_id=1, command="php artisan migrate")
        assert c.status is None
        assert c.duration is None


class TestBackupModel:
    def test_string_duration(self):
        """Forge API returns duration as strings like '18s'."""
        b = Backup(id=1, backup_configuration_id=1, duration="18s")
        assert b.duration == "18s"

    def test_int_duration(self):
        b = Backup(id=1, backup_configuration_id=1, duration=18)
        assert b.duration == 18

    def test_none_duration(self):
        b = Backup(id=1, backup_configuration_id=1, duration=None)
        assert b.duration is None

    def test_string_size(self):
        b = Backup(id=1, backup_configuration_id=1, size="1024")
        assert b.size == "1024"


class TestGitProjectModel:
    def test_minimal(self):
        g = GitProject()
        assert g.repository is None
        assert g.provider is None
