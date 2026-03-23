# Phorge

A keyboard-first terminal UI for managing [Laravel Forge](https://forge.laravel.com) servers, sites, and resources — inspired by [lazygit](https://github.com/jesseduffield/lazygit).

Browse servers, trigger deployments, edit environment files, manage databases, SSH into machines, and more — all without leaving your terminal.

## Features

- **Keyboard-first UX** — lazygit-style three-panel layout with `j/k` navigation, single-key actions, and context-sensitive help
- **Server management** — View server info, SSH keys, daemons, firewall rules, scheduled jobs
- **Site management** — Deployments, deploy scripts, environment files, workers, domains, SSL certificates, commands, git info
- **Database management** — Databases and database users with create/delete
- **SSH integration** — SSH into any server or site with `Ctrl+S`
- **SFTP integration** — Browse files via [termscp](https://github.com/veeso/termscp) with `Ctrl+F`
- **Database tunnel** — Open remote databases in [sqlit](https://github.com/Maxteabag/sqlit) with `Ctrl+D`
- **Environment editor** — Opens `.env` in your preferred editor, detects changes, and uploads automatically
- **Log viewer** — View server/site logs in-app or open in external editor
- **Nicknames** — Assign short aliases to servers/sites, then launch directly with `phorge <nickname>`
- **Quick launch** — Jump straight to a site with `phorge <sitename>` or `phorge <nickname>`
- **Settings modal** — Edit config in-app with `Ctrl+O`
- **Default SSH key** — Configure a default key for quick installation across servers
- **Search/filter** — Press `/` to filter server and site lists in real-time
- **Single binary** — No runtime dependencies, cross-compiled for Linux, macOS, and Windows

## Keyboard Shortcuts

### Navigation

| Key | Action |
|---|---|
| `j` / `k` | Move up / down |
| `g` / `G` | Jump to top / bottom |
| `Tab` / `Shift+Tab` | Cycle panel focus |
| `Enter` | Select / drill in |
| `Esc` | Go back |
| `/` | Search / filter |
| `1`–`9` | Switch section tab |
| `?` | Help |
| `q` | Quit |

### Actions

| Key | Action |
|---|---|
| `Ctrl+S` | SSH to server |
| `Ctrl+F` | SFTP via termscp |
| `Ctrl+D` | Database via sqlit |
| `Ctrl+R` | Refresh |
| `Ctrl+O` | Settings |
| `d` | Deploy site |
| `e` | Edit env / deploy script / open logs in editor |
| `c` | Create resource |
| `x` | Delete resource |
| `r` | Restart (workers, daemons) |
| `n` | Set / remove nickname |
| `D` | Set / clear default server/site |
| `i` | Install default SSH key |
| `l` | View logs |
| `S` | View deploy script |

## Installation

### From source (requires Go 1.22+)

```bash
go install github.com/hinkers/Phorge/cmd/phorge@latest
```

### Build from repo

```bash
git clone https://github.com/hinkers/Phorge.git
cd phorge
make build
```

### From releases

Download the latest binary for your platform from [GitHub Releases](https://github.com/hinkers/Phorge/releases).

## Usage

```bash
phorge                  # launch normally
phorge mysite           # jump straight to a site by name
phorge prod             # jump to a site by nickname
phorge prod --ssh       # SSH into a nicknamed site
phorge prod --sftp      # SFTP into a nicknamed site
phorge prod --db        # open database tunnel for a nicknamed site
phorge --version        # print version
```

Flags can also be used with `.phorge` project defaults (no nickname needed):

```bash
phorge --ssh            # SSH using .phorge default server/site
```

On first launch you'll be prompted for your [Forge API token](https://forge.laravel.com/user-profile/api). The token is saved to `~/.config/phorge/config.toml`.

## Configuration

Config is stored at `~/.config/phorge/config.toml`:

```toml
[forge]
api_key = "your-forge-api-token"
ssh_user = "forge"
default_ssh_key = "~/.ssh/id_ed25519.pub"

[editor]
command = "vim"

[server_users]
"production-1" = "deployer"

[nicknames]
[nicknames.prod]
server = "production-1"
site = "myapp.com"

[nicknames.staging]
server = "staging-1"
site = "staging.myapp.com"
```

| Key | Description | Default |
|---|---|---|
| `forge.api_key` | Forge API token | (required) |
| `forge.ssh_user` | Default SSH username | `forge` |
| `forge.default_ssh_key` | Path to SSH public key for quick install | — |
| `editor.command` | External editor for env/script editing | `vim` |
| `server_users.<name>` | Per-server SSH user override | — |
| `nicknames.<name>` | Short alias mapping to a server/site | — |

## Development

```bash
# Run tests
make test

# Build
make build

# Lint
make vet
```

## Tech Stack

- [Bubbletea](https://github.com/charmbracelet/bubbletea) — TUI framework (Elm Architecture)
- [Bubbles](https://github.com/charmbracelet/bubbles) — TUI components
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Terminal styling
- Go standard library `net/http` — Forge API client

## License

MIT
