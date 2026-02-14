"""CLI entry point for the phorge command."""

from __future__ import annotations


def main() -> None:
    """Launch the Phorge TUI application."""
    from phorge.app import PhorgeApp

    app = PhorgeApp()
    app.run()
