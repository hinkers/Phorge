"""Environment file API endpoints."""

from __future__ import annotations

from phorge.api.client import ForgeClient


class EnvironmentAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def get(self, server_id: int, site_id: int) -> str:
        return await self._client.get_text(
            f"/servers/{server_id}/sites/{site_id}/env"
        )

    async def update(self, server_id: int, site_id: int, content: str) -> None:
        await self._client.put(
            f"/servers/{server_id}/sites/{site_id}/env",
            json={"content": content},
        )
