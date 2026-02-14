"""Backup API endpoints."""

from __future__ import annotations

from typing import Any

from phorge.api.client import ForgeClient
from phorge.api.models import BackupConfig, Backup


class BackupsAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def list_configs(self, server_id: int) -> list[BackupConfig]:
        data = await self._client.get(f"/servers/{server_id}/backup-configs")
        return [BackupConfig(**b) for b in data.get("backups", [])]

    async def get_config(self, server_id: int, config_id: int) -> BackupConfig:
        data = await self._client.get(
            f"/servers/{server_id}/backup-configs/{config_id}"
        )
        return BackupConfig(**data["backup"])

    async def create_config(
        self,
        server_id: int,
        *,
        provider: str,
        credentials: dict[str, Any],
        frequency: str = "daily",
        databases: list[int] | None = None,
        time: str | None = None,
        day_of_week: int | None = None,
    ) -> BackupConfig:
        payload: dict[str, Any] = {
            "provider": provider,
            "credentials": credentials,
            "frequency": frequency,
        }
        if databases:
            payload["databases"] = databases
        if time:
            payload["time"] = time
        if day_of_week is not None:
            payload["day_of_week"] = day_of_week
        data = await self._client.post(
            f"/servers/{server_id}/backup-configs", json=payload
        )
        return BackupConfig(**data["backup"])

    async def delete_config(self, server_id: int, config_id: int) -> None:
        await self._client.delete(f"/servers/{server_id}/backup-configs/{config_id}")

    async def run_backup(self, server_id: int, config_id: int) -> None:
        await self._client.post(f"/servers/{server_id}/backup-configs/{config_id}")

    async def restore_backup(
        self, server_id: int, config_id: int, backup_id: int
    ) -> None:
        await self._client.post(
            f"/servers/{server_id}/backup-configs/{config_id}/backups/{backup_id}"
        )

    async def delete_backup(
        self, server_id: int, config_id: int, backup_id: int
    ) -> None:
        await self._client.delete(
            f"/servers/{server_id}/backup-configs/{config_id}/backups/{backup_id}"
        )
