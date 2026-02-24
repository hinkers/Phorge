package panels

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinkers/Phorge/internal/forge"
	"github.com/hinkers/Phorge/internal/tui/theme"
)

// SiteInfo displays site details as key-value pairs in the detail panel.
type SiteInfo struct {
	site *forge.Site
}

// NewSiteInfo creates a new, empty SiteInfo panel.
func NewSiteInfo() SiteInfo {
	return SiteInfo{}
}

// SetSite replaces the displayed site.
func (s SiteInfo) SetSite(site *forge.Site) SiteInfo {
	s.site = site
	return s
}

// Update handles messages. SiteInfo is mostly display-only.
func (s SiteInfo) Update(msg tea.Msg) (Panel, tea.Cmd) {
	return s, nil
}

// View renders the site info panel as a formatted key-value list.
func (s SiteInfo) View(width, height int, focused bool) string {
	style := theme.InactiveBorderStyle
	titleColor := theme.ColorSubtle
	if focused {
		style = theme.ActiveBorderStyle
		titleColor = theme.ColorPrimary
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Render(" Site ")

	innerWidth := width - 2
	innerHeight := height - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	var lines []string

	if s.site == nil {
		lines = append(lines, theme.NormalItemStyle.Render("No site selected"))
	} else {
		site := s.site
		lines = append(lines, renderInfoKV("Name", site.Name, innerWidth))
		lines = append(lines, renderInfoKV("Directory", site.Directory, innerWidth))
		lines = append(lines, renderInfoKV("Web Dir", site.WebDirectory, innerWidth))
		lines = append(lines, renderInfoKV("Repository", site.Repository, innerWidth))
		lines = append(lines, renderInfoKV("Branch", site.RepositoryBranch, innerWidth))
		lines = append(lines, renderInfoKV("Repo Status", site.RepositoryStatus, innerWidth))
		lines = append(lines, renderInfoKV("PHP", site.PHPVersion, innerWidth))
		lines = append(lines, renderInfoKV("Type", site.ProjectType, innerWidth))
		lines = append(lines, renderStatusKV("Status", site.Status, innerWidth))
		lines = append(lines, renderInfoKV("Quick Deploy", boolToOnOff(site.QuickDeploy), innerWidth))
		lines = append(lines, renderInfoKV("SSL", sslStatus(site.IsSecured), innerWidth))

		// Show aliases if any.
		if len(site.Aliases) > 0 {
			lines = append(lines, "")
			lines = append(lines, theme.LabelStyle.Render("Aliases:"))
			for _, alias := range site.Aliases {
				aliasLine := "  " + theme.ValueStyle.Render(alias)
				lines = append(lines, theme.Truncate(aliasLine, innerWidth))
			}
		}
	}

	// Pad to fill the panel height.
	for len(lines) < innerHeight {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

// HelpBindings returns the key hints for the site info panel.
func (s SiteInfo) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "1-9", Desc: "sections"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}

// boolToOnOff converts a bool to a human-readable "on"/"off" string.
func boolToOnOff(b bool) string {
	if b {
		return "on"
	}
	return "off"
}

// sslStatus converts a bool to "secured"/"not secured".
func sslStatus(secured bool) string {
	if secured {
		return "secured"
	}
	return "not secured"
}
