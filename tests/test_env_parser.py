"""Tests for the .env file parser utility."""

from __future__ import annotations

from phorge.utils.env_parser import parse_env


class TestParseEnv:
    def test_basic_key_value(self):
        result = parse_env("APP_NAME=Laravel")
        assert result == {"APP_NAME": "Laravel"}

    def test_multiple_pairs(self):
        content = "APP_NAME=Laravel\nAPP_ENV=production\nAPP_DEBUG=false"
        result = parse_env(content)
        assert result == {
            "APP_NAME": "Laravel",
            "APP_ENV": "production",
            "APP_DEBUG": "false",
        }

    def test_double_quoted_value(self):
        result = parse_env('DB_PASSWORD="s3cr3t"')
        assert result == {"DB_PASSWORD": "s3cr3t"}

    def test_single_quoted_value(self):
        result = parse_env("DB_PASSWORD='s3cr3t'")
        assert result == {"DB_PASSWORD": "s3cr3t"}

    def test_skips_comments(self):
        content = "# This is a comment\nAPP_NAME=Laravel\n# Another comment"
        result = parse_env(content)
        assert result == {"APP_NAME": "Laravel"}

    def test_skips_blank_lines(self):
        content = "\nAPP_NAME=Laravel\n\n\nAPP_ENV=production\n"
        result = parse_env(content)
        assert result == {"APP_NAME": "Laravel", "APP_ENV": "production"}

    def test_skips_lines_without_equals(self):
        content = "APP_NAME=Laravel\nthis is not a valid line\nAPP_ENV=production"
        result = parse_env(content)
        assert result == {"APP_NAME": "Laravel", "APP_ENV": "production"}

    def test_value_containing_equals(self):
        result = parse_env("DB_PASSWORD=p@ss=word=123")
        assert result == {"DB_PASSWORD": "p@ss=word=123"}

    def test_empty_value(self):
        result = parse_env("DB_PASSWORD=")
        assert result == {"DB_PASSWORD": ""}

    def test_empty_content(self):
        result = parse_env("")
        assert result == {}

    def test_whitespace_stripping(self):
        result = parse_env("  APP_NAME  =  Laravel  ")
        assert result == {"APP_NAME": "Laravel"}

    def test_quoted_value_with_spaces(self):
        result = parse_env('APP_NAME="My App Name"')
        assert result == {"APP_NAME": "My App Name"}

    def test_mismatched_quotes_not_stripped(self):
        result = parse_env("APP_NAME=\"hello'")
        assert result == {"APP_NAME": "\"hello'"}

    def test_single_char_not_stripped_as_quote(self):
        result = parse_env("APP_NAME='")
        assert result == {"APP_NAME": "'"}

    def test_realistic_laravel_env(self):
        content = """APP_NAME=Laravel
APP_ENV=production
APP_KEY=base64:abc123==
APP_DEBUG=false
APP_URL=https://example.com

DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=forge
DB_USERNAME=forge
DB_PASSWORD="s3cr3t#pass@word"

CACHE_DRIVER=redis
QUEUE_CONNECTION=redis
"""
        result = parse_env(content)
        assert result["DB_CONNECTION"] == "mysql"
        assert result["DB_HOST"] == "127.0.0.1"
        assert result["DB_PORT"] == "3306"
        assert result["DB_DATABASE"] == "forge"
        assert result["DB_USERNAME"] == "forge"
        assert result["DB_PASSWORD"] == "s3cr3t#pass@word"
        assert result["APP_KEY"] == "base64:abc123=="
        assert len(result) == 13
