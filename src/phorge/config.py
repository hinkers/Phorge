"""Configuration management for Phorge.

Reads and writes TOML config at ~/.config/phorge/config.toml.
"""

from __future__ import annotations

import os
import tomllib
from dataclasses import dataclass, field
from pathlib import Path

import tomli_w

CONFIG_DIR = Path.home() / ".config" / "phorge"
CONFIG_PATH = CONFIG_DIR / "config.toml"


@dataclass
class ForgeConfig:
    api_key: str = ""


@dataclass
class EditorConfig:
    command: str = "code"


@dataclass
class UIConfig:
    auto_expand_sites: bool = True
    theme: str = "dark"
    vim_keys: bool = False


@dataclass
class PhorgeConfig:
    forge: ForgeConfig = field(default_factory=ForgeConfig)
    editor: EditorConfig = field(default_factory=EditorConfig)
    ui: UIConfig = field(default_factory=UIConfig)


def load_config() -> PhorgeConfig:
    """Load config from TOML file, returning defaults if missing or corrupt."""
    if not CONFIG_PATH.exists():
        return PhorgeConfig()
    try:
        with open(CONFIG_PATH, "rb") as f:
            data = tomllib.load(f)
    except Exception:
        return PhorgeConfig()

    forge_data = data.get("forge", {})
    editor_data = data.get("editor", {})
    ui_data = data.get("ui", {})

    return PhorgeConfig(
        forge=ForgeConfig(
            api_key=forge_data.get("api_key", ""),
        ),
        editor=EditorConfig(
            command=editor_data.get("command", "code"),
        ),
        ui=UIConfig(
            auto_expand_sites=ui_data.get("auto_expand_sites", True),
            theme=ui_data.get("theme", "dark"),
            vim_keys=ui_data.get("vim_keys", False),
        ),
    )


def save_config(config: PhorgeConfig) -> None:
    """Write config to TOML file."""
    CONFIG_DIR.mkdir(parents=True, exist_ok=True)
    os.chmod(CONFIG_DIR, 0o700)

    data = {
        "forge": {
            "api_key": config.forge.api_key,
        },
        "editor": {
            "command": config.editor.command,
        },
        "ui": {
            "auto_expand_sites": config.ui.auto_expand_sites,
            "theme": config.ui.theme,
            "vim_keys": config.ui.vim_keys,
        },
    }

    with open(CONFIG_PATH, "wb") as f:
        tomli_w.dump(data, f)
    os.chmod(CONFIG_PATH, 0o600)


@dataclass
class ProjectConfig:
    server: str | None = None


def load_project_config() -> ProjectConfig:
    """Load project config from .phorge in the current directory."""
    path = Path.cwd() / ".phorge"
    if not path.exists():
        return ProjectConfig()
    try:
        with open(path, "rb") as f:
            data = tomllib.load(f)
    except Exception:
        return ProjectConfig()
    return ProjectConfig(server=data.get("server"))


def save_project_config(config: ProjectConfig) -> None:
    """Write project config to .phorge in the current directory."""
    path = Path.cwd() / ".phorge"
    if config.server is None:
        if path.exists():
            path.unlink()
        return
    with open(path, "wb") as f:
        tomli_w.dump({"server": config.server}, f)


def has_api_key() -> bool:
    """Quick check if an API key is configured."""
    config = load_config()
    return bool(config.forge.api_key)
