"""Git repository API endpoints."""

from __future__ import annotations

from typing import Any

from phorge.api.client import ForgeClient


class GitAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def install(
        self,
        server_id: int,
        site_id: int,
        *,
        provider: str,
        repository: str,
        branch: str = "main",
        composer: bool = True,
    ) -> None:
        payload: dict[str, Any] = {
            "provider": provider,
            "repository": repository,
            "branch": branch,
            "composer": composer,
        }
        await self._client.post(
            f"/servers/{server_id}/sites/{site_id}/git", json=payload
        )

    async def update(
        self, server_id: int, site_id: int, *, branch: str
    ) -> None:
        await self._client.put(
            f"/servers/{server_id}/sites/{site_id}/git",
            json={"branch": branch},
        )

    async def remove(self, server_id: int, site_id: int) -> None:
        await self._client.delete(f"/servers/{server_id}/sites/{site_id}/git")
