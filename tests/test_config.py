"""Tests for the configuration module."""

from __future__ import annotations

from pathlib import Path

import pytest
import tomli_w

from phorge.config import (
    CONFIG_DIR,
    CONFIG_PATH,
    EditorConfig,
    ForgeConfig,
    PhorgeConfig,
    UIConfig,
    has_api_key,
    load_config,
    save_config,
)


@pytest.fixture
def config_dir(tmp_path, monkeypatch):
    """Redirect config to a temp directory."""
    config_dir = tmp_path / ".config" / "phorge"
    config_path = config_dir / "config.toml"
    monkeypatch.setattr("phorge.config.CONFIG_DIR", config_dir)
    monkeypatch.setattr("phorge.config.CONFIG_PATH", config_path)
    return config_dir


@pytest.fixture
def config_path(config_dir):
    return config_dir / "config.toml"


class TestLoadConfig:
    def test_returns_defaults_when_no_file(self, config_dir):
        config = load_config()
        assert config.forge.api_key == ""
        assert config.editor.command == "code"
        assert config.ui.vim_keys is False
        assert config.ui.auto_expand_sites is True
        assert config.ui.theme == "dark"

    def test_loads_existing_config(self, config_dir, config_path):
        config_dir.mkdir(parents=True, exist_ok=True)
        data = {
            "forge": {"api_key": "test-key-123"},
            "editor": {"command": "nvim"},
            "ui": {"vim_keys": True, "theme": "light", "auto_expand_sites": False},
        }
        with open(config_path, "wb") as f:
            tomli_w.dump(data, f)

        config = load_config()
        assert config.forge.api_key == "test-key-123"
        assert config.editor.command == "nvim"
        assert config.ui.vim_keys is True
        assert config.ui.theme == "light"
        assert config.ui.auto_expand_sites is False

    def test_returns_defaults_for_missing_sections(self, config_dir, config_path):
        config_dir.mkdir(parents=True, exist_ok=True)
        data = {"forge": {"api_key": "key-only"}}
        with open(config_path, "wb") as f:
            tomli_w.dump(data, f)

        config = load_config()
        assert config.forge.api_key == "key-only"
        assert config.editor.command == "code"
        assert config.ui.vim_keys is False

    def test_returns_defaults_on_corrupt_file(self, config_dir, config_path):
        config_dir.mkdir(parents=True, exist_ok=True)
        config_path.write_text("this is not valid toml {{{")

        config = load_config()
        assert config.forge.api_key == ""
        assert config.editor.command == "code"


class TestSaveConfig:
    def test_creates_dir_and_writes(self, config_dir, config_path):
        config = PhorgeConfig(
            forge=ForgeConfig(api_key="saved-key"),
            editor=EditorConfig(command="vim"),
            ui=UIConfig(vim_keys=True),
        )
        save_config(config)

        assert config_path.exists()
        loaded = load_config()
        assert loaded.forge.api_key == "saved-key"
        assert loaded.editor.command == "vim"
        assert loaded.ui.vim_keys is True

    def test_roundtrip(self, config_dir, config_path):
        original = PhorgeConfig(
            forge=ForgeConfig(api_key="roundtrip-key"),
            editor=EditorConfig(command="nano"),
            ui=UIConfig(auto_expand_sites=False, theme="light", vim_keys=True),
        )
        save_config(original)
        loaded = load_config()

        assert loaded.forge.api_key == original.forge.api_key
        assert loaded.editor.command == original.editor.command
        assert loaded.ui.auto_expand_sites == original.ui.auto_expand_sites
        assert loaded.ui.theme == original.ui.theme
        assert loaded.ui.vim_keys == original.ui.vim_keys


class TestHasApiKey:
    def test_no_key(self, config_dir):
        assert has_api_key() is False

    def test_with_key(self, config_dir, config_path):
        config_dir.mkdir(parents=True, exist_ok=True)
        data = {"forge": {"api_key": "some-key"}}
        with open(config_path, "wb") as f:
            tomli_w.dump(data, f)
        assert has_api_key() is True

    def test_empty_key(self, config_dir, config_path):
        config_dir.mkdir(parents=True, exist_ok=True)
        data = {"forge": {"api_key": ""}}
        with open(config_path, "wb") as f:
            tomli_w.dump(data, f)
        assert has_api_key() is False
