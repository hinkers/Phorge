"""Pydantic models for Forge API responses."""

from __future__ import annotations

from pydantic import BaseModel, ConfigDict, Field, field_validator


class _ForgeModel(BaseModel):
    """Base model that ignores extra fields from the API."""
    model_config = ConfigDict(extra="ignore")


class User(_ForgeModel):
    id: int
    name: str
    email: str | None = None


def _none_to_list(v: list | None) -> list:
    """Coerce None to empty list for API fields that may return null."""
    return v if v is not None else []


def _none_to_false(v: bool | None) -> bool:
    """Coerce None to False for API fields that may return null."""
    return v if v is not None else False


class Server(_ForgeModel):
    id: int
    name: str
    ip_address: str | None = None
    private_ip_address: str | None = None
    region: str | None = None
    php_version: str | None = None
    php_cli_version: str | None = None
    provider: str | None = None
    type: str | None = None
    status: str | None = None
    is_ready: bool = False
    database_type: str | None = None
    ssh_port: int = 22
    ubuntu_version: str | None = None
    db_status: str | None = None
    redis_status: str | None = None
    network: list[dict] = Field(default_factory=list)
    tags: list[dict] = Field(default_factory=list)

    @field_validator("network", "tags", mode="before")
    @classmethod
    def _coerce_lists(cls, v):
        return _none_to_list(v)

    @field_validator("is_ready", mode="before")
    @classmethod
    def _coerce_bools(cls, v):
        return _none_to_false(v)


class Site(_ForgeModel):
    id: int
    server_id: int | None = None
    name: str
    directory: str | None = None
    repository: str | None = None
    repository_provider: str | None = None
    repository_branch: str | None = None
    repository_status: str | None = None
    quick_deploy: bool = False
    deployment_url: str | None = None
    status: str | None = None
    project_type: str | None = None
    php_version: str | None = None
    app: str | None = None
    wildcards: bool = False
    aliases: list[str] = Field(default_factory=list)
    is_secured: bool = False
    tags: list[dict] = Field(default_factory=list)

    @field_validator("aliases", "tags", mode="before")
    @classmethod
    def _coerce_lists(cls, v):
        return _none_to_list(v)

    @field_validator("quick_deploy", "wildcards", "is_secured", mode="before")
    @classmethod
    def _coerce_bools(cls, v):
        return _none_to_false(v)


class Deployment(_ForgeModel):
    id: int
    server_id: int | None = None
    site_id: int
    type: int | None = None
    commit_hash: str | None = None
    commit_author: str | None = None
    commit_message: str | None = None
    started_at: str | None = None
    ended_at: str | None = None
    status: str | None = None
    displayable_type: str | None = None


class Database(_ForgeModel):
    id: int
    server_id: int | None = None
    name: str
    status: str | None = None
    is_synced: bool = True


class DatabaseUser(_ForgeModel):
    id: int
    server_id: int | None = None
    name: str
    status: str | None = None
    databases: list[int] = Field(default_factory=list)

    @field_validator("databases", mode="before")
    @classmethod
    def _coerce_lists(cls, v):
        return _none_to_list(v)


class SSHKey(_ForgeModel):
    id: int
    name: str
    status: str | None = None


class Daemon(_ForgeModel):
    id: int
    server_id: int | None = None
    command: str
    user: str | None = None
    directory: str | None = None
    processes: int = 1
    startsecs: int = 1
    status: str | None = None


class FirewallRule(_ForgeModel):
    id: int
    server_id: int | None = None
    name: str
    port: int | str | None = None
    ip_address: str | None = None
    type: str | None = None
    status: str | None = None


class ScheduledJob(_ForgeModel):
    id: int
    server_id: int | None = None
    command: str
    user: str | None = None
    frequency: str | None = None
    cron: str | None = None
    status: str | None = None


class Worker(_ForgeModel):
    id: int
    connection: str | None = None
    queue: str | None = None
    timeout: int = 60
    sleep: int = 3
    processes: int = 1
    daemon: bool = True
    force: bool = False
    status: str | None = None

    @field_validator("timeout", mode="before")
    @classmethod
    def _coerce_timeout(cls, v):
        return v if v is not None else 60

    @field_validator("sleep", mode="before")
    @classmethod
    def _coerce_sleep(cls, v):
        return v if v is not None else 3

    @field_validator("processes", mode="before")
    @classmethod
    def _coerce_processes(cls, v):
        return v if v is not None else 1

    @field_validator("daemon", "force", mode="before")
    @classmethod
    def _coerce_bools(cls, v):
        return _none_to_false(v)


class Certificate(_ForgeModel):
    id: int
    domain: str | None = None
    type: str | None = None
    active: bool = False
    status: str | None = None
    existing: bool = False

    @field_validator("active", "existing", mode="before")
    @classmethod
    def _coerce_bools(cls, v):
        return _none_to_false(v)


class BackupConfig(_ForgeModel):
    id: int
    server_id: int | None = None
    day_of_week: int | None = None
    time: str | None = None
    provider: str | None = None
    frequency: str | None = None
    databases: list = Field(default_factory=list)
    backups: list[Backup] | None = None
    backup_time: str | None = None

    @field_validator("databases", mode="before")
    @classmethod
    def _coerce_db_list(cls, v):
        return _none_to_list(v)


class Backup(_ForgeModel):
    id: int
    backup_configuration_id: int
    status: str | None = None
    date: str | None = None
    size: int | str | None = None
    duration: int | str | None = None


# Fix forward reference
BackupConfig.model_rebuild()


class SiteCommand(_ForgeModel):
    id: int
    server_id: int | None = None
    site_id: int
    user_id: int | None = None
    command: str
    status: str | None = None
    created_at: str | None = None
    duration: int | str | None = None
    profile_photo_url: str | None = None
    user_name: str | None = None


class RedirectRule(_ForgeModel):
    id: int
    from_url: str = Field(alias="from", default="")
    to: str = ""
    type: str | None = None
    status: str | None = None


class GitProject(_ForgeModel):
    repository: str | None = None
    provider: str | None = None
    branch: str | None = None
