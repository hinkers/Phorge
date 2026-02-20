"""Parse .env file content into a dictionary."""

from __future__ import annotations


def parse_env(content: str) -> dict[str, str]:
    """Parse .env file content into key=value pairs.

    Handles quoted values, comments, and blank lines.
    """
    result: dict[str, str] = {}
    for line in content.splitlines():
        line = line.strip()
        if not line or line.startswith("#"):
            continue
        if "=" not in line:
            continue
        key, _, value = line.partition("=")
        key = key.strip()
        value = value.strip()
        # Strip surrounding quotes
        if len(value) >= 2 and value[0] == value[-1] and value[0] in ('"', "'"):
            value = value[1:-1]
        result[key] = value
    return result
