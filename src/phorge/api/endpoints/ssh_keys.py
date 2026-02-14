"""SSH key API endpoints."""

from __future__ import annotations

from phorge.api.client import ForgeClient
from phorge.api.models import SSHKey


class SSHKeysAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def list(self, server_id: int) -> list[SSHKey]:
        data = await self._client.get(f"/servers/{server_id}/keys")
        return [SSHKey(**k) for k in data.get("keys", [])]

    async def get(self, server_id: int, key_id: int) -> SSHKey:
        data = await self._client.get(f"/servers/{server_id}/keys/{key_id}")
        return SSHKey(**data["key"])

    async def create(self, server_id: int, name: str, key: str, username: str = "forge") -> SSHKey:
        data = await self._client.post(
            f"/servers/{server_id}/keys",
            json={"name": name, "key": key, "username": username},
        )
        return SSHKey(**data["key"])

    async def delete(self, server_id: int, key_id: int) -> None:
        await self._client.delete(f"/servers/{server_id}/keys/{key_id}")
