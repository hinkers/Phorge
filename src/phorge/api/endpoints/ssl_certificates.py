"""SSL certificate API endpoints."""

from __future__ import annotations

from typing import Any

from phorge.api.client import ForgeClient
from phorge.api.models import Certificate


class SSLCertificatesAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def list(self, server_id: int, site_id: int) -> list[Certificate]:
        data = await self._client.get(
            f"/servers/{server_id}/sites/{site_id}/certificates"
        )
        return [Certificate(**c) for c in data.get("certificates", [])]

    async def get(
        self, server_id: int, site_id: int, certificate_id: int
    ) -> Certificate:
        data = await self._client.get(
            f"/servers/{server_id}/sites/{site_id}/certificates/{certificate_id}"
        )
        return Certificate(**data["certificate"])

    async def obtain_letsencrypt(
        self, server_id: int, site_id: int, domains: list[str]
    ) -> Certificate:
        data = await self._client.post(
            f"/servers/{server_id}/sites/{site_id}/certificates/letsencrypt",
            json={"domains": domains},
        )
        return Certificate(**data["certificate"])

    async def activate(
        self, server_id: int, site_id: int, certificate_id: int
    ) -> None:
        await self._client.post(
            f"/servers/{server_id}/sites/{site_id}/certificates/{certificate_id}/activate"
        )

    async def delete(
        self, server_id: int, site_id: int, certificate_id: int
    ) -> None:
        await self._client.delete(
            f"/servers/{server_id}/sites/{site_id}/certificates/{certificate_id}"
        )
