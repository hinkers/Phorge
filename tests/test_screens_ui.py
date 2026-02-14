"""Textual pilot tests for screens and modals."""

from __future__ import annotations

from unittest.mock import AsyncMock, patch

import pytest

from textual.app import App, ComposeResult
from textual.widgets import Static, Button, Input, Switch, Select

from phorge.api.client import ForgeClient


# ---------------------------------------------------------------------------
# Helpers
# ---------------------------------------------------------------------------

def _make_mock_client() -> ForgeClient:
    client = ForgeClient("test-key")
    client.get = AsyncMock(return_value={})
    client.get_text = AsyncMock(return_value="")
    client.post = AsyncMock(return_value={})
    client.put = AsyncMock(return_value={})
    client.delete = AsyncMock(return_value={})
    return client


# ---------------------------------------------------------------------------
# ConfirmModal
# ---------------------------------------------------------------------------

class TestConfirmModal:
    @pytest.mark.asyncio
    async def test_composes_with_message(self):
        from phorge.screens.confirm import ConfirmModal

        app = App()
        async with app.run_test() as pilot:
            app.push_screen(ConfirmModal("Delete this item?"))
            await pilot.pause()

            screen = app.screen
            message = screen.query_one("#confirm-message", Static)
            assert "Delete this item?" in message.content

            yes_btn = screen.query_one("#confirm-yes", Button)
            assert yes_btn is not None
            no_btn = screen.query_one("#confirm-no", Button)
            assert no_btn is not None

    @pytest.mark.asyncio
    async def test_yes_button_dismisses_true(self):
        from phorge.screens.confirm import ConfirmModal

        results = []

        app = App()
        async with app.run_test() as pilot:
            def capture(result):
                results.append(result)

            modal = ConfirmModal("Proceed?")
            app.push_screen(modal, callback=capture)
            await pilot.pause()

            await pilot.click("#confirm-yes")
            await pilot.pause()

        assert results == [True]

    @pytest.mark.asyncio
    async def test_no_button_dismisses_false(self):
        from phorge.screens.confirm import ConfirmModal

        results = []

        app = App()
        async with app.run_test() as pilot:
            def capture(result):
                results.append(result)

            modal = ConfirmModal("Cancel this?")
            app.push_screen(modal, callback=capture)
            await pilot.pause()

            await pilot.click("#confirm-no")
            await pilot.pause()

        assert results == [False]

    @pytest.mark.asyncio
    async def test_escape_dismisses_false(self):
        from phorge.screens.confirm import ConfirmModal

        results = []

        app = App()
        async with app.run_test() as pilot:
            def capture(result):
                results.append(result)

            modal = ConfirmModal("Quit?")
            app.push_screen(modal, callback=capture)
            await pilot.pause()

            await pilot.press("escape")
            await pilot.pause()

        assert results == [False]


# ---------------------------------------------------------------------------
# InputModal
# ---------------------------------------------------------------------------

class TestInputModal:
    @pytest.mark.asyncio
    async def test_composes_with_title(self):
        from phorge.screens.input_modal import InputModal

        app = App()
        async with app.run_test() as pilot:
            app.push_screen(InputModal("Enter Name", placeholder="my-name"))
            await pilot.pause()

            screen = app.screen
            title = screen.query_one("#input-title", Static)
            assert "Enter Name" in title.content

            inp = screen.query_one("#modal-input", Input)
            assert inp is not None

    @pytest.mark.asyncio
    async def test_ok_button_returns_value(self):
        from phorge.screens.input_modal import InputModal

        results = []

        app = App()
        async with app.run_test() as pilot:
            def capture(result):
                results.append(result)

            modal = InputModal("DB Name")
            app.push_screen(modal, callback=capture)
            await pilot.pause()

            screen = app.screen
            inp = screen.query_one("#modal-input", Input)
            inp.value = "my_database"
            await pilot.click("#input-ok")
            await pilot.pause()

        assert results == ["my_database"]

    @pytest.mark.asyncio
    async def test_cancel_returns_none(self):
        from phorge.screens.input_modal import InputModal

        results = []

        app = App()
        async with app.run_test() as pilot:
            def capture(result):
                results.append(result)

            modal = InputModal("Name")
            app.push_screen(modal, callback=capture)
            await pilot.pause()

            await pilot.click("#input-cancel")
            await pilot.pause()

        assert results == [None]

    @pytest.mark.asyncio
    async def test_escape_returns_none(self):
        from phorge.screens.input_modal import InputModal

        results = []

        app = App()
        async with app.run_test() as pilot:
            def capture(result):
                results.append(result)

            modal = InputModal("Name")
            app.push_screen(modal, callback=capture)
            await pilot.pause()

            await pilot.press("escape")
            await pilot.pause()

        assert results == [None]

    @pytest.mark.asyncio
    async def test_empty_value_returns_none(self):
        from phorge.screens.input_modal import InputModal

        results = []

        app = App()
        async with app.run_test() as pilot:
            def capture(result):
                results.append(result)

            modal = InputModal("Name")
            app.push_screen(modal, callback=capture)
            await pilot.pause()

            # Leave empty and click OK
            await pilot.click("#input-ok")
            await pilot.pause()

        assert results == [None]


# ---------------------------------------------------------------------------
# SSHKeyModal
# ---------------------------------------------------------------------------

class TestSSHKeyModal:
    @pytest.mark.asyncio
    async def test_composes(self):
        from phorge.screens.ssh_key_modal import SSHKeyModal

        app = App()
        async with app.run_test() as pilot:
            app.push_screen(SSHKeyModal())
            await pilot.pause()

            screen = app.screen
            name_input = screen.query_one("#key-name-input", Input)
            assert name_input is not None


# ---------------------------------------------------------------------------
# ConfigScreen
# ---------------------------------------------------------------------------

class TestConfigScreen:
    @pytest.mark.asyncio
    async def test_composes_with_fields(self, tmp_path, monkeypatch):
        from phorge.screens.config_screen import ConfigScreen

        # Redirect config to temp dir so it loads defaults
        config_dir = tmp_path / ".config" / "phorge"
        config_path = config_dir / "config.toml"
        monkeypatch.setattr("phorge.config.CONFIG_DIR", config_dir)
        monkeypatch.setattr("phorge.config.CONFIG_PATH", config_path)

        app = App()
        app.forge_client = _make_mock_client()
        async with app.run_test() as pilot:
            app.push_screen(ConfigScreen())
            await pilot.pause()

            screen = app.screen
            api_key_input = screen.query_one("#cfg-api-key", Input)
            assert api_key_input is not None

            editor_select = screen.query_one("#cfg-editor", Select)
            assert editor_select is not None

            vim_switch = screen.query_one("#cfg-vim-keys", Switch)
            assert vim_switch is not None

            save_btn = screen.query_one("#cfg-save", Button)
            assert save_btn is not None

    @pytest.mark.asyncio
    async def test_save_writes_config(self, tmp_path, monkeypatch):
        from phorge.screens.config_screen import ConfigScreen

        config_dir = tmp_path / ".config" / "phorge"
        config_path = config_dir / "config.toml"
        monkeypatch.setattr("phorge.config.CONFIG_DIR", config_dir)
        monkeypatch.setattr("phorge.config.CONFIG_PATH", config_path)

        results = []
        app = App()
        app.forge_client = _make_mock_client()

        async with app.run_test() as pilot:
            def capture(result):
                results.append(result)

            screen = ConfigScreen()
            app.push_screen(screen, callback=capture)
            await pilot.pause()

            # Set API key
            api_input = app.screen.query_one("#cfg-api-key", Input)
            api_input.value = "new-api-key-123"

            # Click save
            await pilot.click("#cfg-save")
            await pilot.pause()

        # Should dismiss with True
        assert results == [True]
        # Config file should exist
        assert config_path.exists()

    @pytest.mark.asyncio
    async def test_cancel_dismisses_false(self, tmp_path, monkeypatch):
        from phorge.screens.config_screen import ConfigScreen

        config_dir = tmp_path / ".config" / "phorge"
        config_path = config_dir / "config.toml"
        monkeypatch.setattr("phorge.config.CONFIG_DIR", config_dir)
        monkeypatch.setattr("phorge.config.CONFIG_PATH", config_path)

        results = []
        app = App()
        app.forge_client = _make_mock_client()

        async with app.run_test() as pilot:
            def capture(result):
                results.append(result)

            app.push_screen(ConfigScreen(), callback=capture)
            await pilot.pause()

            await pilot.click("#cfg-cancel")
            await pilot.pause()

        assert results == [False]


# ---------------------------------------------------------------------------
# SetupScreen
# ---------------------------------------------------------------------------

class TestSetupScreen:
    @pytest.mark.asyncio
    async def test_composes_with_input(self, tmp_path, monkeypatch):
        from phorge.screens.setup import SetupScreen

        config_dir = tmp_path / ".config" / "phorge"
        config_path = config_dir / "config.toml"
        monkeypatch.setattr("phorge.config.CONFIG_DIR", config_dir)
        monkeypatch.setattr("phorge.config.CONFIG_PATH", config_path)

        app = App()
        async with app.run_test() as pilot:
            app.push_screen(SetupScreen())
            await pilot.pause()

            screen = app.screen
            api_input = screen.query_one("#api-key-input", Input)
            assert api_input is not None
            assert api_input.password is True

            save_btn = screen.query_one("#setup-save-btn", Button)
            assert save_btn is not None

            status = screen.query_one("#setup-status", Static)
            assert status is not None

    @pytest.mark.asyncio
    async def test_empty_key_shows_error(self, tmp_path, monkeypatch):
        from phorge.screens.setup import SetupScreen

        config_dir = tmp_path / ".config" / "phorge"
        config_path = config_dir / "config.toml"
        monkeypatch.setattr("phorge.config.CONFIG_DIR", config_dir)
        monkeypatch.setattr("phorge.config.CONFIG_PATH", config_path)

        app = App()
        async with app.run_test() as pilot:
            app.push_screen(SetupScreen())
            await pilot.pause()

            # Click save without entering a key
            await pilot.click("#setup-save-btn")
            await pilot.pause()

            screen = app.screen
            status = screen.query_one("#setup-status", Static)
            assert "required" in status.content.lower()


# ---------------------------------------------------------------------------
# DetailPanel
# ---------------------------------------------------------------------------

class TestDetailPanel:
    @pytest.mark.asyncio
    async def test_initial_placeholder(self):
        from phorge.widgets.detail_panel import DetailPanel

        class TestApp(App):
            def compose(self) -> ComposeResult:
                yield DetailPanel(id="dp")

        app = TestApp()
        async with app.run_test() as pilot:
            dp = app.query_one(DetailPanel)
            placeholder = dp.query_one("#placeholder", Static)
            assert placeholder is not None
            assert "Select" in placeholder.content

    @pytest.mark.asyncio
    async def test_show_panel_replaces_placeholder(self):
        from phorge.widgets.detail_panel import DetailPanel
        from phorge.widgets.server_tree import NodeData, NodeType

        class TestApp(App):
            def compose(self) -> ComposeResult:
                yield DetailPanel(id="dp")

        app = TestApp()
        app.forge_client = _make_mock_client()
        app.forge_client.get.return_value = {
            "server": {"id": 1, "name": "prod", "ip_address": "1.2.3.4", "ssh_port": 22, "is_ready": True}
        }

        async with app.run_test() as pilot:
            dp = app.query_one(DetailPanel)
            nd = NodeData(NodeType.SERVER_INFO, server_id=1, label="prod", server_ip="1.2.3.4")
            await dp.show_panel(nd)
            await pilot.pause()

            # Placeholder should be gone
            placeholders = dp.query("#placeholder")
            assert len(placeholders) == 0

    @pytest.mark.asyncio
    async def test_show_panel_for_group_type(self):
        from phorge.widgets.detail_panel import DetailPanel
        from phorge.widgets.server_tree import NodeData, NodeType

        class TestApp(App):
            def compose(self) -> ComposeResult:
                yield DetailPanel(id="dp")

        app = TestApp()
        async with app.run_test() as pilot:
            dp = app.query_one(DetailPanel)
            nd = NodeData(NodeType.SERVER_ROOT, server_id=1, label="server")
            await dp.show_panel(nd)
            await pilot.pause()

            # Should show "No panel" fallback since SERVER_ROOT has no panel
            # The current panel should be a Static with "No panel" message


# ---------------------------------------------------------------------------
# MainScreen (integration)
# ---------------------------------------------------------------------------

class TestMainScreenCompose:
    @pytest.mark.asyncio
    async def test_composes_tree_and_detail(self, tmp_path, monkeypatch):
        """MainScreen should compose with a ServerTree and DetailPanel."""
        from phorge.screens.main import MainScreen
        from phorge.widgets.server_tree import ServerTree
        from phorge.widgets.detail_panel import DetailPanel

        config_dir = tmp_path / ".config" / "phorge"
        config_path = config_dir / "config.toml"
        monkeypatch.setattr("phorge.config.CONFIG_DIR", config_dir)
        monkeypatch.setattr("phorge.config.CONFIG_PATH", config_path)

        app = App()
        app.forge_client = _make_mock_client()
        app.forge_client.get.return_value = {"servers": []}

        async with app.run_test() as pilot:
            app.push_screen(MainScreen())
            await pilot.pause()

            screen = app.screen
            tree = screen.query_one(ServerTree)
            assert tree is not None

            detail = screen.query_one(DetailPanel)
            assert detail is not None

    @pytest.mark.asyncio
    async def test_loads_servers_on_mount(self, tmp_path, monkeypatch):
        """MainScreen should call the servers API on mount."""
        from phorge.screens.main import MainScreen
        from phorge.widgets.server_tree import ServerTree

        config_dir = tmp_path / ".config" / "phorge"
        config_path = config_dir / "config.toml"
        monkeypatch.setattr("phorge.config.CONFIG_DIR", config_dir)
        monkeypatch.setattr("phorge.config.CONFIG_PATH", config_path)

        app = App()
        app.forge_client = _make_mock_client()
        app.forge_client.get.return_value = {
            "servers": [
                {"id": 1, "name": "web-1", "ip_address": "10.0.0.1", "ssh_port": 22, "is_ready": True},
                {"id": 2, "name": "web-2", "ip_address": "10.0.0.2", "ssh_port": 22, "is_ready": True},
            ]
        }

        async with app.run_test() as pilot:
            app.push_screen(MainScreen())
            await pilot.pause()

            screen = app.screen
            tree = screen.query_one(ServerTree)
            # Tree should have been populated with 2 servers
            assert len(tree.root.children) == 2
            app.forge_client.get.assert_called()
