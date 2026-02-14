"""Tests for the Forge API client."""

from __future__ import annotations

from unittest.mock import AsyncMock, MagicMock, patch

import httpx
import pytest

from phorge.api.client import BASE_URL, ForgeClient
from phorge.api.exceptions import (
    ForgeAPIError,
    ForgeAuthenticationError,
    ForgeNotFoundError,
    ForgeRateLimitError,
    ForgeValidationError,
)


@pytest.fixture
def client():
    return ForgeClient("test-key")


class TestClientInit:
    def test_stores_api_key(self, client):
        assert client._api_key == "test-key"

    def test_lazy_client_starts_none(self, client):
        assert client._client is None

    def test_get_client_creates_client(self, client):
        http_client = client._get_client()
        assert isinstance(http_client, httpx.AsyncClient)
        assert str(http_client.base_url).rstrip("/") == BASE_URL.rstrip("/")

    def test_get_client_reuses_client(self, client):
        c1 = client._get_client()
        c2 = client._get_client()
        assert c1 is c2


class TestErrorHandling:
    def _make_response(self, status_code, json_data=None, text=""):
        response = MagicMock(spec=httpx.Response)
        response.status_code = status_code
        response.is_success = 200 <= status_code < 300
        response.text = text
        response.url = "https://forge.laravel.com/api/v1/test"
        if json_data is not None:
            response.json.return_value = json_data
        else:
            response.json.side_effect = Exception("no json")
        return response

    def test_success_no_error(self, client):
        response = self._make_response(200)
        client._handle_errors(response)  # should not raise

    def test_401_raises_auth_error(self, client):
        response = self._make_response(401)
        with pytest.raises(ForgeAuthenticationError, match="Invalid API key"):
            client._handle_errors(response)

    def test_404_raises_not_found(self, client):
        response = self._make_response(404)
        with pytest.raises(ForgeNotFoundError, match="Resource not found"):
            client._handle_errors(response)

    def test_422_raises_validation_error(self, client):
        response = self._make_response(422, json_data={"errors": {"name": ["required"]}})
        with pytest.raises(ForgeValidationError):
            client._handle_errors(response)

    def test_422_with_no_json(self, client):
        response = self._make_response(422, text="validation failed")
        with pytest.raises(ForgeValidationError):
            client._handle_errors(response)

    def test_429_raises_rate_limit(self, client):
        response = self._make_response(429)
        with pytest.raises(ForgeRateLimitError, match="Rate limit"):
            client._handle_errors(response)

    def test_500_raises_generic_error(self, client):
        response = self._make_response(500, json_data={"message": "Server error"})
        with pytest.raises(ForgeAPIError, match="500.*Server error"):
            client._handle_errors(response)

    def test_500_with_no_json(self, client):
        response = self._make_response(500, text="Internal Server Error")
        with pytest.raises(ForgeAPIError, match="500"):
            client._handle_errors(response)


class TestRequest:
    @pytest.mark.asyncio
    async def test_get_returns_json(self, client):
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.is_success = True
        mock_response.content = b'{"servers": []}'
        mock_response.json.return_value = {"servers": []}

        mock_http = AsyncMock()
        mock_http.request.return_value = mock_response
        mock_http.is_closed = False
        client._client = mock_http

        result = await client.get("/servers")
        assert result == {"servers": []}
        mock_http.request.assert_called_once_with("GET", "/servers", json=None, params=None)

    @pytest.mark.asyncio
    async def test_get_text_returns_plain_text(self, client):
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.is_success = True
        mock_response.text = "#!/bin/bash\ncd /home/forge"

        mock_http = AsyncMock()
        mock_http.request.return_value = mock_response
        mock_http.is_closed = False
        client._client = mock_http

        result = await client.get_text("/servers/1/sites/1/deployment/script")
        assert result == "#!/bin/bash\ncd /home/forge"

    @pytest.mark.asyncio
    async def test_204_returns_empty_dict(self, client):
        mock_response = MagicMock()
        mock_response.status_code = 204
        mock_response.is_success = True
        mock_response.content = b""

        mock_http = AsyncMock()
        mock_http.request.return_value = mock_response
        mock_http.is_closed = False
        client._client = mock_http

        result = await client.post("/servers/1/reboot")
        assert result == {}

    @pytest.mark.asyncio
    async def test_empty_content_returns_empty_dict(self, client):
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_response.is_success = True
        mock_response.content = b""

        mock_http = AsyncMock()
        mock_http.request.return_value = mock_response
        mock_http.is_closed = False
        client._client = mock_http

        result = await client.get("/empty")
        assert result == {}


class TestClose:
    @pytest.mark.asyncio
    async def test_close_when_no_client(self, client):
        await client.close()  # should not raise

    @pytest.mark.asyncio
    async def test_close_calls_aclose(self, client):
        mock_http = AsyncMock()
        mock_http.is_closed = False
        client._client = mock_http

        await client.close()
        mock_http.aclose.assert_called_once()

    @pytest.mark.asyncio
    async def test_close_when_already_closed(self, client):
        mock_http = AsyncMock()
        mock_http.is_closed = True
        client._client = mock_http

        await client.close()
        mock_http.aclose.assert_not_called()
