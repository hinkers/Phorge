"""Log API endpoints."""

from __future__ import annotations

from phorge.api.client import ForgeClient


class LogsAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def get_server_logs(self, server_id: int) -> str:
        data = await self._client.get(f"/servers/{server_id}/logs")
        return data.get("content", "")

    async def get_site_logs(self, server_id: int, site_id: int) -> str:
        data = await self._client.get(
            f"/servers/{server_id}/sites/{site_id}/logs"
        )
        return data.get("content", "")

    async def clear_site_logs(self, server_id: int, site_id: int) -> None:
        await self._client.delete(
            f"/servers/{server_id}/sites/{site_id}/logs"
        )
