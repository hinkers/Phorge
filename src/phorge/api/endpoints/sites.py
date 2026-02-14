"""Site API endpoints."""

from __future__ import annotations

from phorge.api.client import ForgeClient
from phorge.api.models import Site


class SitesAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def list(self, server_id: int) -> list[Site]:
        data = await self._client.get(f"/servers/{server_id}/sites")
        return [Site(**s) for s in data.get("sites", [])]

    async def get(self, server_id: int, site_id: int) -> Site:
        data = await self._client.get(f"/servers/{server_id}/sites/{site_id}")
        return Site(**data["site"])

    async def update_aliases(self, server_id: int, site_id: int, aliases: list[str]) -> Site:
        data = await self._client.put(
            f"/servers/{server_id}/sites/{site_id}/aliases",
            json={"aliases": aliases},
        )
        return Site(**data["site"])

    async def change_php_version(self, server_id: int, site_id: int, version: str) -> None:
        await self._client.put(
            f"/servers/{server_id}/sites/{site_id}/php",
            json={"version": version},
        )
