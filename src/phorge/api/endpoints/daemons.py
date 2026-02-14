"""Daemon API endpoints."""

from __future__ import annotations

from typing import Any

from phorge.api.client import ForgeClient
from phorge.api.models import Daemon


class DaemonsAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def list(self, server_id: int) -> list[Daemon]:
        data = await self._client.get(f"/servers/{server_id}/daemons")
        return [Daemon(**d) for d in data.get("daemons", [])]

    async def get(self, server_id: int, daemon_id: int) -> Daemon:
        data = await self._client.get(f"/servers/{server_id}/daemons/{daemon_id}")
        return Daemon(**data["daemon"])

    async def create(
        self,
        server_id: int,
        *,
        command: str,
        user: str = "forge",
        directory: str | None = None,
        processes: int = 1,
        startsecs: int = 1,
    ) -> Daemon:
        payload: dict[str, Any] = {
            "command": command,
            "user": user,
            "processes": processes,
            "startsecs": startsecs,
        }
        if directory:
            payload["directory"] = directory
        data = await self._client.post(f"/servers/{server_id}/daemons", json=payload)
        return Daemon(**data["daemon"])

    async def restart(self, server_id: int, daemon_id: int) -> None:
        await self._client.post(f"/servers/{server_id}/daemons/{daemon_id}/restart")

    async def delete(self, server_id: int, daemon_id: int) -> None:
        await self._client.delete(f"/servers/{server_id}/daemons/{daemon_id}")
