"""Worker API endpoints."""

from __future__ import annotations

from typing import Any

from phorge.api.client import ForgeClient
from phorge.api.models import Worker


class WorkersAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def list(self, server_id: int, site_id: int) -> list[Worker]:
        data = await self._client.get(
            f"/servers/{server_id}/sites/{site_id}/workers"
        )
        return [Worker(**w) for w in data.get("workers", [])]

    async def get(self, server_id: int, site_id: int, worker_id: int) -> Worker:
        data = await self._client.get(
            f"/servers/{server_id}/sites/{site_id}/workers/{worker_id}"
        )
        return Worker(**data["worker"])

    async def create(
        self,
        server_id: int,
        site_id: int,
        *,
        connection: str = "redis",
        queue: str = "default",
        timeout: int = 60,
        sleep: int = 3,
        processes: int = 1,
        daemon: bool = True,
        force: bool = False,
        php_version: str | None = None,
    ) -> Worker:
        payload: dict[str, Any] = {
            "connection": connection,
            "queue": queue,
            "timeout": timeout,
            "sleep": sleep,
            "processes": processes,
            "daemon": daemon,
            "force": force,
        }
        if php_version:
            payload["php_version"] = php_version
        data = await self._client.post(
            f"/servers/{server_id}/sites/{site_id}/workers",
            json=payload,
        )
        return Worker(**data["worker"])

    async def restart(self, server_id: int, site_id: int, worker_id: int) -> None:
        await self._client.post(
            f"/servers/{server_id}/sites/{site_id}/workers/{worker_id}/restart"
        )

    async def delete(self, server_id: int, site_id: int, worker_id: int) -> None:
        await self._client.delete(
            f"/servers/{server_id}/sites/{site_id}/workers/{worker_id}"
        )

    async def get_output(self, server_id: int, site_id: int, worker_id: int) -> str:
        data = await self._client.get(
            f"/servers/{server_id}/sites/{site_id}/workers/{worker_id}/output"
        )
        return data.get("output", "")
