"""Async HTTP client for the Laravel Forge API."""

from __future__ import annotations

from typing import Any

import httpx

from phorge.api.exceptions import (
    ForgeAPIError,
    ForgeAuthenticationError,
    ForgeNotFoundError,
    ForgeRateLimitError,
    ForgeValidationError,
)

BASE_URL = "https://forge.laravel.com/api/v1"


class ForgeClient:
    """Async API client for Laravel Forge.

    Uses a single long-lived httpx.AsyncClient to reuse TCP/TLS connections.
    The client is lazily initialized on first request.
    """

    def __init__(self, api_key: str) -> None:
        self._api_key = api_key
        self._client: httpx.AsyncClient | None = None

    @property
    def api_key(self) -> str:
        return self._api_key

    def _get_client(self) -> httpx.AsyncClient:
        if self._client is None or self._client.is_closed:
            self._client = httpx.AsyncClient(
                base_url=BASE_URL,
                headers={
                    "Authorization": f"Bearer {self._api_key}",
                    "Accept": "application/json",
                    "Content-Type": "application/json",
                },
                timeout=httpx.Timeout(30.0, connect=10.0),
            )
        return self._client

    async def _request(
        self,
        method: str,
        path: str,
        *,
        json: dict[str, Any] | None = None,
        params: dict[str, Any] | None = None,
    ) -> dict[str, Any]:
        client = self._get_client()
        response = await client.request(method, path, json=json, params=params)
        self._handle_errors(response)
        if response.status_code == 204:
            return {}
        if not response.content:
            return {}
        return response.json()

    def _handle_errors(self, response: httpx.Response) -> None:
        if response.is_success:
            return
        if response.status_code == 401:
            raise ForgeAuthenticationError("Invalid API key")
        if response.status_code == 404:
            raise ForgeNotFoundError(f"Resource not found: {response.url}")
        if response.status_code == 422:
            try:
                details = response.json()
            except Exception:
                details = {"error": response.text}
            raise ForgeValidationError(details)
        if response.status_code == 429:
            raise ForgeRateLimitError("Rate limit exceeded")
        try:
            error_body = response.json()
            msg = error_body.get("message", response.text)
        except Exception:
            msg = response.text
        raise ForgeAPIError(f"API error {response.status_code}: {msg}")

    async def get(self, path: str, **kwargs: Any) -> dict[str, Any]:
        return await self._request("GET", path, **kwargs)

    async def get_text(self, path: str) -> str:
        """GET request that returns the response body as plain text."""
        client = self._get_client()
        response = await client.request("GET", path)
        self._handle_errors(response)
        return response.text

    async def post(self, path: str, **kwargs: Any) -> dict[str, Any]:
        return await self._request("POST", path, **kwargs)

    async def put(self, path: str, **kwargs: Any) -> dict[str, Any]:
        return await self._request("PUT", path, **kwargs)

    async def delete(self, path: str, **kwargs: Any) -> dict[str, Any]:
        return await self._request("DELETE", path, **kwargs)

    async def close(self) -> None:
        if self._client and not self._client.is_closed:
            await self._client.aclose()
