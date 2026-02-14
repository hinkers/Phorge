"""Custom exceptions for the Forge API client."""

from __future__ import annotations


class ForgeAPIError(Exception):
    """Base exception for Forge API errors."""


class ForgeAuthenticationError(ForgeAPIError):
    """401 - Invalid or missing API key."""


class ForgeNotFoundError(ForgeAPIError):
    """404 - Resource not found."""


class ForgeValidationError(ForgeAPIError):
    """422 - Validation error with details."""

    def __init__(self, details: dict) -> None:
        self.details = details
        super().__init__(str(details))


class ForgeRateLimitError(ForgeAPIError):
    """429 - Rate limit exceeded."""
