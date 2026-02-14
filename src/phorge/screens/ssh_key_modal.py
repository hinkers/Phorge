"""SSH key input modal for adding keys to servers."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Horizontal, Vertical
from textual.screen import ModalScreen
from textual.widgets import Static, Button, Input, TextArea


class SSHKeyModal(ModalScreen[tuple[str, str] | None]):
    """Modal that collects SSH key name and content."""

    DEFAULT_CSS = """
    SSHKeyModal {
        align: center middle;
    }
    #ssh-key-dialog {
        width: 80;
        height: auto;
        max-height: 25;
        border: thick $background 80%;
        background: $surface;
        padding: 1 2;
    }
    #ssh-key-dialog Static {
        margin-bottom: 1;
    }
    #key-name-input {
        width: 100%;
        margin-bottom: 1;
    }
    #key-content-input {
        width: 100%;
        height: 8;
        margin-bottom: 1;
    }
    #ssh-key-buttons {
        width: 100%;
        height: 3;
        align: center middle;
    }
    #ssh-key-buttons Button {
        margin: 0 1;
    }
    """

    BINDINGS = [
        ("escape", "cancel", "Cancel"),
    ]

    def compose(self) -> ComposeResult:
        with Vertical(id="ssh-key-dialog"):
            yield Static("[bold]Add SSH Key[/bold]")
            yield Static("Key Name:")
            yield Input(placeholder="My SSH Key", id="key-name-input")
            yield Static("Public Key Content:")
            yield TextArea(id="key-content-input")
            with Horizontal(id="ssh-key-buttons"):
                yield Button("Add Key", variant="primary", id="ssh-add")
                yield Button("Cancel", variant="default", id="ssh-cancel")

    def on_mount(self) -> None:
        self.query_one("#key-name-input", Input).focus()

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "ssh-add":
            name = self.query_one("#key-name-input", Input).value.strip()
            content = self.query_one("#key-content-input", TextArea).text.strip()
            if name and content:
                self.dismiss((name, content))
            else:
                self.notify("Both name and key content are required", severity="error")
        else:
            self.dismiss(None)

    def action_cancel(self) -> None:
        self.dismiss(None)
