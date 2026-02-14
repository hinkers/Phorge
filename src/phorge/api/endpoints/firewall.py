"""Firewall rule API endpoints."""

from __future__ import annotations

from typing import Any

from phorge.api.client import ForgeClient
from phorge.api.models import FirewallRule


class FirewallAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def list(self, server_id: int) -> list[FirewallRule]:
        data = await self._client.get(f"/servers/{server_id}/firewall-rules")
        return [FirewallRule(**r) for r in data.get("rules", [])]

    async def get(self, server_id: int, rule_id: int) -> FirewallRule:
        data = await self._client.get(f"/servers/{server_id}/firewall-rules/{rule_id}")
        return FirewallRule(**data["rule"])

    async def create(
        self,
        server_id: int,
        *,
        name: str,
        port: int | str,
        ip_address: str | None = None,
        type: str = "allow",
    ) -> FirewallRule:
        payload: dict[str, Any] = {
            "name": name,
            "port": port,
            "type": type,
        }
        if ip_address:
            payload["ip_address"] = ip_address
        data = await self._client.post(
            f"/servers/{server_id}/firewall-rules", json=payload
        )
        return FirewallRule(**data["rule"])

    async def delete(self, server_id: int, rule_id: int) -> None:
        await self._client.delete(f"/servers/{server_id}/firewall-rules/{rule_id}")
