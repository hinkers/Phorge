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
from phorge.config import load_config, load_project_config, save_project_config, ProjectConfig


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
    #picker-hint {
        width: 100%;
        text-align: center;
        color: $text-muted;
        margin-top: 1;
    }
    """

    BINDINGS = [
        ("escape", "cancel", "Cancel"),
        ("ctrl+d", "set_default", "Set Default"),
    ]

    def __init__(self, default_server: str | None = None, **kwargs) -> None:
        super().__init__(**kwargs)
        self._servers: list[Server] = []
        self._default_server = default_server

    def compose(self) -> ComposeResult:
        with Vertical(id="picker-dialog"):
            yield Static("Select a Server", id="picker-title")
            yield LoadingIndicator(id="picker-loading")
            yield OptionList(id="picker-list")
            yield Static("ctrl+d toggle default  * = default", id="picker-hint")

    async def on_mount(self) -> None:
        option_list = self.query_one(OptionList)
        option_list.display = False
        config = load_config()
        if config.ui.vim_keys:
            option_list._bindings.bind("j", "cursor_down", "Down", show=False)
            option_list._bindings.bind("k", "cursor_up", "Up", show=False)
        await self._fetch_servers()

    def _format_option(self, server: Server, is_default: bool) -> str:
        ip = server.ip_address or "no ip"
        region = server.region or "N/A"
        provider = server.provider or "N/A"
        marker = " *" if is_default else ""
        return f"{server.name}  ({ip})  {region} / {provider}{marker}"

    async def _fetch_servers(self) -> None:
        loading = self.query_one(LoadingIndicator)
        option_list = self.query_one(OptionList)
        project = load_project_config()
        self._project_default = project.server
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
                is_default = (
                    self._project_default is not None
                    and server.name.lower() == self._project_default.lower()
                )
                option_list.add_option(
                    Option(self._format_option(server, is_default))
                )
            if self._default_server and self._servers:
                for server in self._servers:
                    if server.name.lower() == self._default_server.lower():
                        self.dismiss(server)
                        return
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

    def action_set_default(self) -> None:
        option_list = self.query_one(OptionList)
        index = option_list.highlighted
        if index is None or not (0 <= index < len(self._servers)):
            return
        server = self._servers[index]
        if self._project_default and server.name.lower() == self._project_default.lower():
            save_project_config(ProjectConfig(server=None))
            self._project_default = None
            self.notify(f"Cleared default server", markup=False)
        else:
            save_project_config(ProjectConfig(server=server.name))
            self._project_default = server.name
            self.notify(f"Set {server.name} as default server", markup=False)
        option_list.clear_options()
        for i, s in enumerate(self._servers):
            is_default = (
                self._project_default is not None
                and s.name.lower() == self._project_default.lower()
            )
            option_list.add_option(Option(self._format_option(s, is_default)))
        option_list.highlighted = index

    def action_cancel(self) -> None:
        self.dismiss(None)
