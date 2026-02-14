# Phorge

A terminal UI for managing [Laravel Forge](https://forge.laravel.com) servers, sites, and resources — built with [Textual](https://github.com/Textualize/textual).

Browse servers, trigger deployments, edit environment files, manage databases, SSH into machines, and more — all without leaving your terminal.

## Features

- **Server management** — View server info, SSH keys, daemons, firewall rules, scheduled jobs
- **Site management** — Deployments, deployment scripts, environment files, workers, backups, domains, SSL certificates, commands, git repository info
- **Database management** — Databases and database users per server
- **Tree navigation** — Two-pane layout with a collapsible tree on the left and detail panels on the right
- **Lazy loading** — Sites are fetched on-demand when you expand a server node
- **SSH integration** — SSH directly into any server or site with one keystroke
- **Environment editor** — Opens `.env` in your preferred editor, detects changes, and uploads automatically
- **Command palette** — Fuzzy search across all servers and actions (`Ctrl+P`)
- **Vim keybindings** — Optional vim-style navigation (`h/j/k/l`) in the tree
- **Configurable** — API key, editor command, and UI preferences stored in `~/.config/phorge/config.toml`

## Keyboard Shortcuts

| Key | Action |
|---|---|
| `Ctrl+Q` | Quit |
| `Ctrl+P` | Command palette |
| `Ctrl+S` | SSH to selected server |
| `Ctrl+R` | Refresh current view |
| `Ctrl+E` | Edit configuration |

## Installation

Requires Python 3.11+.

```bash
# With pipx (recommended)
pipx install .

# With pip
pip install .

# For development
poetry install
```

## Usage

```bash
phorge
```

On first launch you'll be prompted for your [Forge API token](https://forge.laravel.com/user-profile/api). The token is saved to `~/.config/phorge/config.toml`.

## Configuration

Config is stored at `~/.config/phorge/config.toml`:

```toml
[forge]
api_key = "your-forge-api-token"

[editor]
command = "code"  # or "nvim", "vim", "nano", etc.

[ui]
vim_keys = false
```

## Development

```bash
# Install dev dependencies
poetry install

# Run tests
pytest

# Run with textual dev console
textual run --dev -c phorge
```

## Tech Stack

- [Textual](https://github.com/Textualize/textual) — TUI framework
- [httpx](https://github.com/encode/httpx) — Async HTTP client
- [Pydantic](https://docs.pydantic.dev) — API response validation
- [Poetry](https://python-poetry.org) — Dependency management

## License

MIT
