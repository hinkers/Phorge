"""Scheduled job API endpoints."""

from __future__ import annotations

from typing import Any

from phorge.api.client import ForgeClient
from phorge.api.models import ScheduledJob


class ScheduledJobsAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def list(self, server_id: int) -> list[ScheduledJob]:
        data = await self._client.get(f"/servers/{server_id}/jobs")
        return [ScheduledJob(**j) for j in data.get("jobs", [])]

    async def get(self, server_id: int, job_id: int) -> ScheduledJob:
        data = await self._client.get(f"/servers/{server_id}/jobs/{job_id}")
        return ScheduledJob(**data["job"])

    async def create(
        self,
        server_id: int,
        *,
        command: str,
        frequency: str = "nightly",
        user: str = "forge",
    ) -> ScheduledJob:
        payload: dict[str, Any] = {
            "command": command,
            "frequency": frequency,
            "user": user,
        }
        data = await self._client.post(f"/servers/{server_id}/jobs", json=payload)
        return ScheduledJob(**data["job"])

    async def delete(self, server_id: int, job_id: int) -> None:
        await self._client.delete(f"/servers/{server_id}/jobs/{job_id}")

    async def get_output(self, server_id: int, job_id: int) -> str:
        data = await self._client.get(f"/servers/{server_id}/jobs/{job_id}/output")
        return data.get("output", "")
