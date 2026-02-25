package panels

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinkers/Phorge/internal/forge"
	"github.com/hinkers/Phorge/internal/tui/theme"
)

// ServerInfo displays server details as key-value pairs in the detail panel.
type ServerInfo struct {
	server *forge.Server
}

// NewServerInfo creates a new, empty ServerInfo panel.
func NewServerInfo() ServerInfo {
	return ServerInfo{}
}

// SetServer replaces the displayed server.
func (s ServerInfo) SetServer(srv *forge.Server) ServerInfo {
	s.server = srv
	return s
}

// Update handles messages. ServerInfo is mostly display-only.
func (s ServerInfo) Update(msg tea.Msg) (Panel, tea.Cmd) {
	return s, nil
}

// View renders the server info panel as a formatted key-value list.
func (s ServerInfo) View(width, height int, focused bool) string {
	style := theme.InactiveBorderStyle
	titleColor := theme.ColorSubtle
	if focused {
		style = theme.ActiveBorderStyle
		titleColor = theme.ColorPrimary
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Render(" Server ")

	innerWidth := width - 2
	innerHeight := height - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	var lines []string

	if s.server == nil {
		lines = append(lines, theme.NormalItemStyle.Render("No server selected"))
	} else {
		srv := s.server
		lines = append(lines, renderInfoKV("Name", srv.Name, innerWidth))
		lines = append(lines, renderInfoKV("IP", srv.IPAddress, innerWidth))
		lines = append(lines, renderInfoKV("Private IP", srv.PrivateIPAddress, innerWidth))
		lines = append(lines, renderInfoKV("Provider", srv.Provider, innerWidth))
		lines = append(lines, renderInfoKV("Region", srv.Region, innerWidth))
		lines = append(lines, renderInfoKV("PHP", srv.PHPVersion, innerWidth))
		lines = append(lines, renderInfoKV("Ubuntu", srv.UbuntuVersion, innerWidth))
		lines = append(lines, renderInfoKV("DB Type", srv.DatabaseType, innerWidth))
		lines = append(lines, renderInfoKV("DB Status", srv.DBStatus, innerWidth))
		lines = append(lines, renderInfoKV("Redis", srv.RedisStatus, innerWidth))
		lines = append(lines, renderStatusKV("Status", srv.Status, innerWidth))
		lines = append(lines, renderInfoKV("SSH Port", fmt.Sprintf("%d", srv.SSHPort), innerWidth))
	}

	// Pad to fill the panel height.
	for len(lines) < innerHeight-1 {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

// HelpBindings returns the key hints for the server info panel.
func (s ServerInfo) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}

// renderInfoKV renders a label-value pair for the info panels.
func renderInfoKV(label, value string, maxWidth int) string {
	if value == "" {
		value = "-"
	}
	l := theme.LabelStyle.Render(label + ":")
	v := theme.ValueStyle.Render(value)
	line := l + " " + v
	return theme.Truncate(line, maxWidth)
}

// renderStatusKV renders a status value with colour based on its content.
func renderStatusKV(label, value string, maxWidth int) string {
	if value == "" {
		value = "-"
	}
	l := theme.LabelStyle.Render(label + ":")

	var v string
	switch strings.ToLower(value) {
	case "active", "installed":
		v = theme.ActiveStatusStyle.Render(value)
	default:
		v = theme.ErrorStatusStyle.Render(value)
	}

	line := l + " " + v
	return theme.Truncate(line, maxWidth)
}
