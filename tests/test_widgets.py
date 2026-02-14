"""Tests for widgets: ServerTree, DetailPanel, NodeData."""

from __future__ import annotations

import pytest

from phorge.api.models import Server, Site
from phorge.widgets.server_tree import NodeData, NodeType, ServerTree
from phorge.widgets.detail_panel import DetailPanel, _get_panel_class


class TestNodeData:
    def test_defaults(self):
        nd = NodeData(NodeType.SERVER_INFO, server_id=1)
        assert nd.site_id is None
        assert nd.label == ""
        assert nd.server_ip is None
        assert nd.ssh_port == 22
        assert nd.loaded is False

    def test_full(self):
        nd = NodeData(
            NodeType.DEPLOYMENTS,
            server_id=1,
            site_id=10,
            label="example.com",
            server_ip="1.2.3.4",
            ssh_port=2222,
            loaded=True,
        )
        assert nd.node_type == NodeType.DEPLOYMENTS
        assert nd.server_id == 1
        assert nd.site_id == 10
        assert nd.ssh_port == 2222
        assert nd.loaded is True


class TestNodeType:
    def test_all_types_exist(self):
        expected = [
            "SERVER_ROOT", "SERVER_INFO", "SITES_GROUP", "SITE_ROOT", "SITE_INFO",
            "DEPLOYMENTS", "DEPLOYMENT_SCRIPT", "LOGS", "ENVIRONMENT", "WORKERS",
            "BACKUPS", "DOMAINS", "DATABASES_SITE", "SSL_CERTIFICATES", "COMMANDS",
            "GIT_REPOSITORY", "SSH_KEYS", "DAEMONS", "FIREWALL_RULES",
            "SCHEDULED_JOBS", "DATABASES_SERVER", "DATABASE_USERS",
        ]
        for name in expected:
            assert hasattr(NodeType, name), f"Missing NodeType.{name}"


class TestPanelRegistry:
    """Verify that the panel registry maps all expected node types."""

    def test_all_panel_types_have_class(self):
        paneled_types = [
            NodeType.SERVER_INFO, NodeType.SITE_INFO, NodeType.DEPLOYMENTS,
            NodeType.DEPLOYMENT_SCRIPT, NodeType.LOGS, NodeType.ENVIRONMENT,
            NodeType.WORKERS, NodeType.BACKUPS, NodeType.DOMAINS,
            NodeType.DATABASES_SERVER, NodeType.DATABASES_SITE,
            NodeType.DATABASE_USERS, NodeType.SSL_CERTIFICATES,
            NodeType.COMMANDS, NodeType.GIT_REPOSITORY, NodeType.SSH_KEYS,
            NodeType.DAEMONS, NodeType.FIREWALL_RULES, NodeType.SCHEDULED_JOBS,
        ]
        for nt in paneled_types:
            cls = _get_panel_class(nt)
            assert cls is not None, f"No panel class for {nt.name}"

    def test_group_types_return_none(self):
        for nt in [NodeType.SERVER_ROOT, NodeType.SITE_ROOT, NodeType.SITES_GROUP]:
            assert _get_panel_class(nt) is None


class TestExceptions:
    """Test that exception classes work correctly."""

    def test_forge_api_error(self):
        from phorge.api.exceptions import ForgeAPIError
        e = ForgeAPIError("test")
        assert str(e) == "test"

    def test_validation_error_stores_details(self):
        from phorge.api.exceptions import ForgeValidationError
        details = {"errors": {"name": ["required"]}}
        e = ForgeValidationError(details)
        assert e.details == details
        assert "name" in str(e)

    def test_exception_hierarchy(self):
        from phorge.api.exceptions import (
            ForgeAPIError,
            ForgeAuthenticationError,
            ForgeNotFoundError,
            ForgeRateLimitError,
            ForgeValidationError,
        )
        assert issubclass(ForgeAuthenticationError, ForgeAPIError)
        assert issubclass(ForgeNotFoundError, ForgeAPIError)
        assert issubclass(ForgeValidationError, ForgeAPIError)
        assert issubclass(ForgeRateLimitError, ForgeAPIError)
