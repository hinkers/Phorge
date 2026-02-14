"""Site command API endpoints."""

from __future__ import annotations

from phorge.api.client import ForgeClient
from phorge.api.models import SiteCommand


class CommandsAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def list(self, server_id: int, site_id: int) -> list[SiteCommand]:
        data = await self._client.get(
            f"/servers/{server_id}/sites/{site_id}/commands"
        )
        return [SiteCommand(**c) for c in data.get("commands", [])]

    async def get(self, server_id: int, site_id: int, command_id: int) -> SiteCommand:
        data = await self._client.get(
            f"/servers/{server_id}/sites/{site_id}/commands/{command_id}"
        )
        return SiteCommand(**data["command"])

    async def execute(self, server_id: int, site_id: int, command: str) -> SiteCommand:
        data = await self._client.post(
            f"/servers/{server_id}/sites/{site_id}/commands",
            json={"command": command},
        )
        return SiteCommand(**data["command"])
