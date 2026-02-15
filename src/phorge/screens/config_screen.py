"""Configuration editing screen."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Horizontal, Vertical
from textual.screen import ModalScreen
from textual.widgets import Static, Button, Input, Select, Switch

from phorge.config import load_config, save_config


class ConfigScreen(ModalScreen[bool]):
    """Modal for editing Phorge configuration."""

    DEFAULT_CSS = """
    ConfigScreen {
        align: center middle;
    }
    #config-dialog {
        width: 70;
        height: auto;
        max-height: 25;
        border: thick $background 80%;
        background: $surface;
        padding: 1 2;
    }
    #config-dialog Static {
        margin-bottom: 0;
    }
    .config-field {
        width: 100%;
        margin-bottom: 1;
    }
    .switch-row {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    .switch-row Static {
        width: 1fr;
        padding-top: 1;
    }
    .switch-row Switch {
        width: auto;
    }
    #config-buttons {
        width: 100%;
        height: 3;
        align: center middle;
        margin-top: 1;
    }
    #config-buttons Button {
        margin: 0 1;
    }
    """

    BINDINGS = [
        ("escape", "cancel", "Cancel"),
    ]

    def compose(self) -> ComposeResult:
        config = load_config()

        with Vertical(id="config-dialog"):
            yield Static("[bold]Phorge Configuration[/bold]")

            yield Static("Forge API Key:")
            yield Input(
                value=config.forge.api_key,
                password=True,
                id="cfg-api-key",
                classes="config-field",
            )

            yield Static("Default Editor:")
            yield Select(
                [
                    ("VS Code", "code"),
                    ("Neovim", "nvim"),
                    ("Vim", "vim"),
                    ("Nano", "nano"),
                ],
                value=config.editor.command,
                id="cfg-editor",
                classes="config-field",
            )

            with Horizontal(classes="switch-row"):
                yield Static("Vim Keybindings:")
                yield Switch(value=config.ui.vim_keys, id="cfg-vim-keys")

            with Horizontal(id="config-buttons"):
                yield Button("Save", variant="primary", id="cfg-save")
                yield Button("Cancel", variant="default", id="cfg-cancel")

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "cfg-save":
            self._save()
        else:
            self.dismiss(False)

    def _save(self) -> None:
        config = load_config()

        api_key = self.query_one("#cfg-api-key", Input).value.strip()
        editor_select = self.query_one("#cfg-editor", Select)
        editor = str(editor_select.value) if editor_select.value != Select.BLANK else config.editor.command

        vim_keys = self.query_one("#cfg-vim-keys", Switch).value

        config.forge.api_key = api_key
        config.editor.command = editor
        config.ui.vim_keys = vim_keys

        save_config(config)

        # Update the app's forge client if API key changed
        if api_key and api_key != self.app.forge_client.api_key:
            from phorge.api.client import ForgeClient
            self.app.forge_client = ForgeClient(api_key)

        self.notify("Configuration saved")
        self.dismiss(True)

    def action_cancel(self) -> None:
        self.dismiss(False)
