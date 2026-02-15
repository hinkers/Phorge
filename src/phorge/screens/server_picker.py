"""Server picker modal for selecting a single server."""

from __future__ import annotations

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.screen import ModalScreen
from textual.widgets import LoadingIndicator, OptionList, Static
from textual.widgets.option_list import Option

from phorge.api.endpoints.servers import ServersAPI
from phorge.api.exceptions import ForgeAPIError
from phorge.api.models import Server
from phorge.config import load_config


class ServerPicker(ModalScreen[Server | None]):
    """Modal that lists servers and returns the selected one."""

    DEFAULT_CSS = """
    ServerPicker {
        align: center middle;
    }
    #picker-dialog {
        width: 70;
        height: auto;
        max-height: 80%;
        border: thick $background 80%;
        background: $surface;
        padding: 1 2;
    }
    #picker-title {
        width: 100%;
        text-align: center;
        text-style: bold;
        margin-bottom: 1;
    }
    #picker-list {
        width: 100%;
        height: auto;
        max-height: 20;
    }
    #picker-loading {
        width: 100%;
        height: 3;
    }
    """

    BINDINGS = [
        ("escape", "cancel", "Cancel"),
    ]

    def __init__(self, **kwargs) -> None:
        super().__init__(**kwargs)
        self._servers: list[Server] = []

    def compose(self) -> ComposeResult:
        with Vertical(id="picker-dialog"):
            yield Static("Select a Server", id="picker-title")
            yield LoadingIndicator(id="picker-loading")
            yield OptionList(id="picker-list")

    async def on_mount(self) -> None:
        option_list = self.query_one(OptionList)
        option_list.display = False
        config = load_config()
        if config.ui.vim_keys:
            option_list._bindings.bind("j", "cursor_down", "Down", show=False)
            option_list._bindings.bind("k", "cursor_up", "Up", show=False)
        await self._fetch_servers()

    async def _fetch_servers(self) -> None:
        loading = self.query_one(LoadingIndicator)
        option_list = self.query_one(OptionList)
        try:
            client = getattr(self.app, "forge_client", None)
            if client is None:
                self.notify("No API client available", severity="error")
                self.dismiss(None)
                return
            api = ServersAPI(client)
            self._servers = await api.list()
            loading.display = False
            option_list.display = True
            for server in self._servers:
                ip = server.ip_address or "no ip"
                region = server.region or "N/A"
                provider = server.provider or "N/A"
                option_list.add_option(
                    Option(f"{server.name}  ({ip})  {region} / {provider}")
                )
            if self._servers:
                option_list.focus()
        except ForgeAPIError as e:
            self.notify(f"Error loading servers: {e}", severity="error", markup=False)
            self.dismiss(None)
        except Exception as e:
            self.notify(f"Unexpected error: {e}", severity="error", markup=False)
            self.dismiss(None)

    def on_option_list_option_selected(self, event: OptionList.OptionSelected) -> None:
        index = event.option_index
        if 0 <= index < len(self._servers):
            self.dismiss(self._servers[index])

    def action_cancel(self) -> None:
        self.dismiss(None)
