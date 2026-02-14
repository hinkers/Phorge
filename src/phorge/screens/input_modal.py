"""Generic text input modal dialog."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Horizontal, Vertical
from textual.screen import ModalScreen
from textual.widgets import Static, Button, Input


class InputModal(ModalScreen[str | None]):
    """Modal that collects a single text input."""

    DEFAULT_CSS = """
    InputModal {
        align: center middle;
    }
    #input-dialog {
        width: 60;
        height: auto;
        max-height: 15;
        border: thick $background 80%;
        background: $surface;
        padding: 1 2;
    }
    #input-title {
        width: 100%;
        margin-bottom: 1;
    }
    #modal-input {
        width: 100%;
        margin-bottom: 1;
    }
    #input-buttons {
        width: 100%;
        height: 3;
        align: center middle;
    }
    #input-buttons Button {
        margin: 0 1;
    }
    """

    BINDINGS = [
        ("escape", "cancel", "Cancel"),
    ]

    def __init__(self, title: str, placeholder: str = "", **kwargs) -> None:
        super().__init__(**kwargs)
        self.title_text = title
        self.placeholder = placeholder

    def compose(self) -> ComposeResult:
        with Vertical(id="input-dialog"):
            yield Static(f"[bold]{self.title_text}[/bold]", id="input-title")
            yield Input(placeholder=self.placeholder, id="modal-input")
            with Horizontal(id="input-buttons"):
                yield Button("OK", variant="primary", id="input-ok")
                yield Button("Cancel", variant="default", id="input-cancel")

    def on_mount(self) -> None:
        self.query_one("#modal-input", Input).focus()

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "input-ok":
            value = self.query_one("#modal-input", Input).value.strip()
            self.dismiss(value if value else None)
        else:
            self.dismiss(None)

    def on_input_submitted(self, event: Input.Submitted) -> None:
        value = event.value.strip()
        self.dismiss(value if value else None)

    def action_cancel(self) -> None:
        self.dismiss(None)
