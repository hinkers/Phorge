"""Server API endpoints."""

from __future__ import annotations

from phorge.api.client import ForgeClient
from phorge.api.models import Server, User


class ServersAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def list(self) -> list[Server]:
        data = await self._client.get("/servers")
        return [Server(**s) for s in data.get("servers", [])]

    async def get(self, server_id: int) -> Server:
        data = await self._client.get(f"/servers/{server_id}")
        return Server(**data["server"])

    async def reboot(self, server_id: int) -> None:
        await self._client.post(f"/servers/{server_id}/reboot")

    async def get_user(self) -> User:
        data = await self._client.get("/user")
        return User(**data["user"])
