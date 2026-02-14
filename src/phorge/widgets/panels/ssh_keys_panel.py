"""SSH keys panel for managing server SSH keys."""

from __future__ import annotations

from pathlib import Path

from textual.app import ComposeResult
from textual.containers import Vertical
from textual.widgets import DataTable, Static, Button
from textual import work

from phorge.api.endpoints.ssh_keys import SSHKeysAPI
from phorge.widgets.server_tree import NodeData


class SSHKeysPanel(Vertical):
    """Shows SSH keys and allows adding new ones."""

    DEFAULT_CSS = """
    SSHKeysPanel {
        height: 1fr;
    }
    SSHKeysPanel .action-bar {
        height: 3;
        layout: horizontal;
        margin-bottom: 1;
    }
    SSHKeysPanel .action-bar Button {
        margin: 0 1 0 0;
    }
    SSHKeysPanel DataTable {
        height: 1fr;
    }
    """

    def __init__(self, node_data: NodeData, **kwargs) -> None:
        super().__init__(**kwargs)
        self.node_data = node_data

    def compose(self) -> ComposeResult:
        yield Static("[bold]SSH Keys[/bold]", classes="panel-title")
        with Vertical(classes="action-bar"):
            yield Button("Add Key", id="btn-add", variant="primary")
            yield Button("Add from ~/.ssh", id="btn-add-local", variant="default")
            yield Button("Refresh", id="btn-refresh", variant="default")
        yield DataTable(id="ssh-keys-table", cursor_type="row")

    def on_mount(self) -> None:
        table = self.query_one(DataTable)
        table.add_columns("ID", "Name", "Status")
        self.load_data()

    @work(exclusive=True)
    async def load_data(self) -> None:
        table = self.query_one(DataTable)
        table.loading = True
        try:
            api = SSHKeysAPI(self.app.forge_client)
            keys = await api.list(self.node_data.server_id)
            table.clear()
            for k in keys:
                table.add_row(
                    str(k.id),
                    k.name,
                    k.status or "",
                    key=str(k.id),
                )
            if not keys:
                self.notify("No SSH keys found", severity="information")
        except Exception as e:
            self.notify(f"Error: {e}", severity="error", markup=False)
        finally:
            table.loading = False

    def on_button_pressed(self, event: Button.Pressed) -> None:
        if event.button.id == "btn-refresh":
            self.load_data()
        elif event.button.id == "btn-add":
            self._add_key()
        elif event.button.id == "btn-add-local":
            self._add_local_key()

    @work
    async def _add_key(self) -> None:
        from phorge.screens.ssh_key_modal import SSHKeyModal

        result = await self.app.push_screen_wait(SSHKeyModal())
        if result:
            name, key_content = result
            api = SSHKeysAPI(self.app.forge_client)
            await api.create(self.node_data.server_id, name, key_content)
            self.notify(f"SSH key '{name}' added")
            self.load_data()

    @work
    async def _add_local_key(self) -> None:
        """Read the default SSH public key from ~/.ssh and add it."""
        ssh_dir = Path.home() / ".ssh"
        pub_key_path = ssh_dir / "id_ed25519.pub"
        if not pub_key_path.exists():
            pub_key_path = ssh_dir / "id_rsa.pub"
        if not pub_key_path.exists():
            self.notify("No SSH public key found in ~/.ssh", severity="error")
            return

        key_content = pub_key_path.read_text().strip()

        from phorge.screens.input_modal import InputModal

        name = await self.app.push_screen_wait(
            InputModal("Key Name", placeholder="My Key")
        )
        if name:
            api = SSHKeysAPI(self.app.forge_client)
            await api.create(self.node_data.server_id, name, key_content)
            self.notify(f"SSH key '{name}' added from {pub_key_path.name}")
            self.load_data()

    def on_data_table_row_selected(self, event: DataTable.RowSelected) -> None:
        key_id = int(str(event.row_key.value))
        self._confirm_delete_key(key_id)

    @work
    async def _confirm_delete_key(self, key_id: int) -> None:
        from phorge.screens.confirm import ConfirmModal

        confirmed = await self.app.push_screen_wait(
            ConfirmModal(f"Delete SSH key #{key_id}?")
        )
        if confirmed:
            api = SSHKeysAPI(self.app.forge_client)
            await api.delete(self.node_data.server_id, key_id)
            self.notify("SSH key deleted")
            self.load_data()
