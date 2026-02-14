"""Shared test fixtures for Phorge."""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock

import pytest

from phorge.api.client import ForgeClient
from phorge.api.models import Server, Site, Deployment, Database, SSHKey
from phorge.widgets.server_tree import NodeData, NodeType


@pytest.fixture
def mock_client():
    """A ForgeClient with mocked HTTP methods."""
    client = ForgeClient("test-api-key")
    client.get = AsyncMock()
    client.get_text = AsyncMock()
    client.post = AsyncMock(return_value={})
    client.put = AsyncMock(return_value={})
    client.delete = AsyncMock(return_value={})
    return client


@pytest.fixture
def sample_server_data():
    """Raw server API response data."""
    return {
        "id": 1,
        "name": "production",
        "ip_address": "1.2.3.4",
        "private_ip_address": "10.0.0.1",
        "region": "us-east-1",
        "php_version": "8.2",
        "php_cli_version": "8.2",
        "provider": "ocean2",
        "type": "app",
        "status": "installed",
        "is_ready": True,
        "database_type": "mysql8",
        "ssh_port": 22,
        "ubuntu_version": "22.04",
        "db_status": "installed",
        "redis_status": "installed",
        "network": [],
        "tags": [],
    }


@pytest.fixture
def sample_site_data():
    """Raw site API response data."""
    return {
        "id": 10,
        "server_id": 1,
        "name": "example.com",
        "directory": "/home/forge/example.com",
        "repository": "user/repo",
        "repository_provider": "github",
        "repository_branch": "main",
        "repository_status": "installed",
        "quick_deploy": True,
        "status": "installed",
        "project_type": "php",
        "php_version": "php82",
        "wildcards": False,
        "aliases": ["www.example.com"],
        "is_secured": True,
        "tags": [],
    }


@pytest.fixture
def sample_server(sample_server_data):
    return Server(**sample_server_data)


@pytest.fixture
def sample_site(sample_site_data):
    return Site(**sample_site_data)


@pytest.fixture
def server_node_data():
    """NodeData for a server."""
    return NodeData(
        node_type=NodeType.SERVER_INFO,
        server_id=1,
        label="production",
        server_ip="1.2.3.4",
        ssh_port=22,
    )


@pytest.fixture
def site_node_data():
    """NodeData for a site."""
    return NodeData(
        node_type=NodeType.DEPLOYMENTS,
        server_id=1,
        site_id=10,
        label="example.com",
    )
