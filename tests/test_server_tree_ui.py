"""Textual pilot tests for the ServerTree widget."""

from __future__ import annotations

import pytest

from textual.app import App, ComposeResult
from textual.widgets import Tree

from phorge.api.models import Server, Site
from phorge.widgets.server_tree import NodeData, NodeType, ServerTree


class TreeTestApp(App):
    """Minimal app host for testing the ServerTree."""

    CSS = "Screen { layout: vertical; }"

    def compose(self) -> ComposeResult:
        yield ServerTree(id="tree")


class TestServerTreePopulation:
    @pytest.mark.asyncio
    async def test_empty_tree(self):
        app = TreeTestApp()
        async with app.run_test() as pilot:
            tree = app.query_one(ServerTree)
            tree.populate_servers([])
            await pilot.pause()
            # Root should be expanded but have no children
            assert len(tree.root.children) == 0

    @pytest.mark.asyncio
    async def test_single_server(self):
        app = TreeTestApp()
        async with app.run_test() as pilot:
            tree = app.query_one(ServerTree)
            server = Server(id=1, name="production", ip_address="1.2.3.4", ssh_port=22, is_ready=True)
            tree.populate_servers([server])
            await pilot.pause()

            # Root should have 1 child (server node)
            assert len(tree.root.children) == 1
            server_node = tree.root.children[0]

            # Server node data should be SERVER_ROOT type
            assert server_node.data.node_type == NodeType.SERVER_ROOT
            assert server_node.data.server_id == 1
            assert server_node.data.server_ip == "1.2.3.4"
            assert server_node.data.ssh_port == 22

    @pytest.mark.asyncio
    async def test_server_children_structure(self):
        app = TreeTestApp()
        async with app.run_test() as pilot:
            tree = app.query_one(ServerTree)
            server = Server(id=1, name="web-1", ip_address="10.0.0.1")
            tree.populate_servers([server])
            await pilot.pause()

            server_node = tree.root.children[0]
            children = server_node.children

            # Should have: Server Info, Sites, SSH Keys, Daemons, Firewall Rules,
            # Scheduled Jobs, Databases, Database Users = 8 children
            assert len(children) == 8

            # Verify types
            child_types = [c.data.node_type for c in children]
            assert NodeType.SERVER_INFO in child_types
            assert NodeType.SITES_GROUP in child_types
            assert NodeType.SSH_KEYS in child_types
            assert NodeType.DAEMONS in child_types
            assert NodeType.FIREWALL_RULES in child_types
            assert NodeType.SCHEDULED_JOBS in child_types
            assert NodeType.DATABASES_SERVER in child_types
            assert NodeType.DATABASE_USERS in child_types

    @pytest.mark.asyncio
    async def test_multiple_servers(self):
        app = TreeTestApp()
        async with app.run_test() as pilot:
            tree = app.query_one(ServerTree)
            servers = [
                Server(id=1, name="web-1", ip_address="10.0.0.1"),
                Server(id=2, name="web-2", ip_address="10.0.0.2"),
                Server(id=3, name="db-1", ip_address="10.0.0.3"),
            ]
            tree.populate_servers(servers)
            await pilot.pause()

            assert len(tree.root.children) == 3

            # Each server has its own ID
            ids = [c.data.server_id for c in tree.root.children]
            assert ids == [1, 2, 3]

    @pytest.mark.asyncio
    async def test_repopulate_clears_old(self):
        app = TreeTestApp()
        async with app.run_test() as pilot:
            tree = app.query_one(ServerTree)

            # First population
            tree.populate_servers([Server(id=1, name="old-server")])
            await pilot.pause()
            assert len(tree.root.children) == 1

            # Re-populate should replace
            tree.populate_servers([
                Server(id=2, name="new-1"),
                Server(id=3, name="new-2"),
            ])
            await pilot.pause()
            assert len(tree.root.children) == 2
            assert tree.root.children[0].data.server_id == 2


class TestServerTreeSites:
    @pytest.mark.asyncio
    async def test_add_sites_to_node(self):
        app = TreeTestApp()
        async with app.run_test() as pilot:
            tree = app.query_one(ServerTree)
            server = Server(id=1, name="web-1", ip_address="10.0.0.1")
            tree.populate_servers([server])
            await pilot.pause()

            # Find the SITES_GROUP node
            server_node = tree.root.children[0]
            sites_node = None
            for child in server_node.children:
                if child.data.node_type == NodeType.SITES_GROUP:
                    sites_node = child
                    break
            assert sites_node is not None

            # Add sites
            sites = [
                Site(id=10, server_id=1, name="example.com"),
                Site(id=11, server_id=1, name="api.example.com"),
            ]
            tree.add_sites_to_node(sites_node, sites)
            await pilot.pause()

            # Sites node should now have 2 children
            assert len(sites_node.children) == 2
            assert sites_node.data.loaded is True

    @pytest.mark.asyncio
    async def test_site_node_children(self):
        app = TreeTestApp()
        async with app.run_test() as pilot:
            tree = app.query_one(ServerTree)
            server = Server(id=1, name="web-1")
            tree.populate_servers([server])
            await pilot.pause()

            server_node = tree.root.children[0]
            sites_node = [c for c in server_node.children if c.data.node_type == NodeType.SITES_GROUP][0]

            sites = [Site(id=10, server_id=1, name="example.com")]
            tree.add_sites_to_node(sites_node, sites)
            await pilot.pause()

            site_node = sites_node.children[0]
            assert site_node.data.node_type == NodeType.SITE_ROOT
            assert site_node.data.site_id == 10

            # Each site should have 12 sub-items
            assert len(site_node.children) == 12
            child_types = [c.data.node_type for c in site_node.children]
            assert NodeType.SITE_INFO in child_types
            assert NodeType.DEPLOYMENTS in child_types
            assert NodeType.DEPLOYMENT_SCRIPT in child_types
            assert NodeType.LOGS in child_types
            assert NodeType.ENVIRONMENT in child_types
            assert NodeType.WORKERS in child_types
            assert NodeType.BACKUPS in child_types
            assert NodeType.DOMAINS in child_types
            assert NodeType.DATABASES_SITE in child_types
            assert NodeType.SSL_CERTIFICATES in child_types
            assert NodeType.COMMANDS in child_types
            assert NodeType.GIT_REPOSITORY in child_types


class TestServerTreeVimKeys:
    @pytest.mark.asyncio
    async def test_enable_vim_keys(self):
        app = TreeTestApp()
        async with app.run_test() as pilot:
            tree = app.query_one(ServerTree)
            # Should not raise
            tree.enable_vim_keys()

    @pytest.mark.asyncio
    async def test_cursor_parent_action(self):
        app = TreeTestApp()
        async with app.run_test() as pilot:
            tree = app.query_one(ServerTree)
            server = Server(id=1, name="web-1")
            tree.populate_servers([server])
            await pilot.pause()

            # action_cursor_parent should not raise even with no cursor
            tree.action_cursor_parent()


class TestServerTreeNodeData:
    @pytest.mark.asyncio
    async def test_server_ip_propagates(self):
        """Server IP and SSH port should propagate to child nodes."""
        app = TreeTestApp()
        async with app.run_test() as pilot:
            tree = app.query_one(ServerTree)
            server = Server(id=1, name="web-1", ip_address="192.168.1.1", ssh_port=2222)
            tree.populate_servers([server])
            await pilot.pause()

            server_node = tree.root.children[0]
            # Server root should have IP
            assert server_node.data.server_ip == "192.168.1.1"
            assert server_node.data.ssh_port == 2222

            # Server Info child should also have IP
            info_node = [c for c in server_node.children if c.data.node_type == NodeType.SERVER_INFO][0]
            assert info_node.data.server_ip == "192.168.1.1"
            assert info_node.data.ssh_port == 2222

    @pytest.mark.asyncio
    async def test_server_no_ip(self):
        """Server without IP should show 'no ip'."""
        app = TreeTestApp()
        async with app.run_test() as pilot:
            tree = app.query_one(ServerTree)
            server = Server(id=1, name="web-1")  # no ip_address
            tree.populate_servers([server])
            await pilot.pause()

            server_node = tree.root.children[0]
            assert server_node.data.server_ip is None
