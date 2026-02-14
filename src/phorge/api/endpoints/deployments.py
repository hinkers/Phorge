"""Deployment API endpoints."""

from __future__ import annotations

from phorge.api.client import ForgeClient
from phorge.api.models import Deployment


class DeploymentsAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def list(self, server_id: int, site_id: int) -> list[Deployment]:
        data = await self._client.get(
            f"/servers/{server_id}/sites/{site_id}/deployment-history"
        )
        return [Deployment(**d) for d in data.get("deployments", [])]

    async def get(self, server_id: int, site_id: int, deployment_id: int) -> Deployment:
        data = await self._client.get(
            f"/servers/{server_id}/sites/{site_id}/deployment-history/{deployment_id}"
        )
        return Deployment(**data["deployment"])

    async def get_output(self, server_id: int, site_id: int, deployment_id: int) -> str:
        data = await self._client.get(
            f"/servers/{server_id}/sites/{site_id}/deployment-history/{deployment_id}/output"
        )
        return data.get("output", "")

    async def deploy(self, server_id: int, site_id: int) -> None:
        await self._client.post(
            f"/servers/{server_id}/sites/{site_id}/deployment/deploy"
        )

    async def get_log(self, server_id: int, site_id: int) -> str:
        data = await self._client.get(
            f"/servers/{server_id}/sites/{site_id}/deployment/log"
        )
        return data.get("output", "")

    async def get_script(self, server_id: int, site_id: int) -> str:
        return await self._client.get_text(
            f"/servers/{server_id}/sites/{site_id}/deployment/script"
        )

    async def update_script(self, server_id: int, site_id: int, content: str) -> None:
        await self._client.put(
            f"/servers/{server_id}/sites/{site_id}/deployment/script",
            json={"content": content},
        )

    async def enable_quick_deploy(self, server_id: int, site_id: int) -> None:
        await self._client.post(f"/servers/{server_id}/sites/{site_id}/deployment")

    async def disable_quick_deploy(self, server_id: int, site_id: int) -> None:
        await self._client.delete(f"/servers/{server_id}/sites/{site_id}/deployment")

    async def reset_status(self, server_id: int, site_id: int) -> None:
        await self._client.post(
            f"/servers/{server_id}/sites/{site_id}/deployment/reset"
        )
