"""First-run setup screen for API key configuration."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical, Center
from textual.screen import Screen
from textual.widgets import Static, Button, Input
from textual import work

from rich.markup import escape

from phorge.api.client import ForgeClient
from phorge.api.exceptions import ForgeAuthenticationError
from phorge.config import load_config, save_config


class SetupScreen(Screen):
    """First-run screen to collect and validate the Forge API key."""

    DEFAULT_CSS = """
    SetupScreen {
        align: center middle;
    }
    #setup-container {
        width: 70;
        height: auto;
        max-height: 20;
        border: thick $primary;
        background: $surface;
        padding: 2 4;
    }
    #setup-container Static {
        margin-bottom: 1;
    }
    #api-key-input {
        width: 100%;
        margin-bottom: 1;
    }
    #setup-save-btn {
        width: 100%;
    }
    #setup-status {
        height: 1;
        margin-top: 1;
    }
    """

    def compose(self) -> ComposeResult:
        with Center():
            with Vertical(id="setup-container"):
                yield Static("[bold]Welcome to Phorge[/bold]")
                yield Static("Enter your Laravel Forge API key to get started.")
                yield Static("[dim]https://forge.laravel.com/user-profile/api[/dim]")
                yield Input(
                    placeholder="Your Forge API key...",
                    password=True,
                    id="api-key-input",
                )
                yield Button("Save & Continue", variant="primary", id="setup-save-btn")
                yield Static("", id="setup-status")

    def on_mount(self) -> None:
        self.query_one("#api-key-input", Input).focus()

    def on_input_submitted(self, event: Input.Submitted) -> None:
        self._save_and_validate()

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "setup-save-btn":
            self._save_and_validate()

    @work(exclusive=True)
    async def _save_and_validate(self) -> None:
        api_key = self.query_one("#api-key-input", Input).value.strip()
        status = self.query_one("#setup-status", Static)

        if not api_key:
            status.update("[red]API key is required[/red]")
            return

        status.update("[dim]Validating API key...[/dim]")

        client = ForgeClient(api_key)
        try:
            await client.get("/user")
        except ForgeAuthenticationError:
            status.update("[red]Invalid API key. Please try again.[/red]")
            await client.close()
            return
        except Exception as e:
            status.update(f"[red]Connection error: {escape(str(e))}[/red]")
            await client.close()
            return

        config = load_config()
        config.forge.api_key = api_key
        save_config(config)

        self.app.forge_client = client

        from phorge.screens.main import MainScreen
        self.app.switch_screen(MainScreen())
