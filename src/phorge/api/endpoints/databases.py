"""Database and database user API endpoints."""

from __future__ import annotations

from typing import Any

from phorge.api.client import ForgeClient
from phorge.api.models import Database, DatabaseUser


class DatabasesAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def list(self, server_id: int) -> list[Database]:
        data = await self._client.get(f"/servers/{server_id}/databases")
        return [Database(**d) for d in data.get("databases", [])]

    async def get(self, server_id: int, database_id: int) -> Database:
        data = await self._client.get(f"/servers/{server_id}/databases/{database_id}")
        return Database(**data["database"])

    async def create(self, server_id: int, name: str, user: str | None = None, password: str | None = None) -> Database:
        payload: dict[str, Any] = {"name": name}
        if user:
            payload["user"] = user
        if password:
            payload["password"] = password
        data = await self._client.post(f"/servers/{server_id}/databases", json=payload)
        return Database(**data["database"])

    async def delete(self, server_id: int, database_id: int) -> None:
        await self._client.delete(f"/servers/{server_id}/databases/{database_id}")

    async def sync(self, server_id: int) -> None:
        await self._client.post(f"/servers/{server_id}/databases/sync")


class DatabaseUsersAPI:
    def __init__(self, client: ForgeClient) -> None:
        self._client = client

    async def list(self, server_id: int) -> list[DatabaseUser]:
        data = await self._client.get(f"/servers/{server_id}/database-users")
        return [DatabaseUser(**u) for u in data.get("users", [])]

    async def get(self, server_id: int, user_id: int) -> DatabaseUser:
        data = await self._client.get(f"/servers/{server_id}/database-users/{user_id}")
        return DatabaseUser(**data["user"])

    async def create(self, server_id: int, name: str, password: str, databases: list[int] | None = None) -> DatabaseUser:
        payload: dict[str, Any] = {"name": name, "password": password}
        if databases:
            payload["databases"] = databases
        data = await self._client.post(f"/servers/{server_id}/database-users", json=payload)
        return DatabaseUser(**data["user"])

    async def update(self, server_id: int, user_id: int, databases: list[int]) -> DatabaseUser:
        data = await self._client.put(
            f"/servers/{server_id}/database-users/{user_id}",
            json={"databases": databases},
        )
        return DatabaseUser(**data["user"])

    async def delete(self, server_id: int, user_id: int) -> None:
        await self._client.delete(f"/servers/{server_id}/database-users/{user_id}")
