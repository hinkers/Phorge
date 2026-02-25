package panels

import (
	"strings"

	tea "charm.land/bubbletea/v2"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinkers/Phorge/internal/forge"
	"github.com/hinkers/Phorge/internal/tui/theme"
)

// GitPanel shows repository information for a site as key-value pairs.
// This is a read-only info panel -- no API calls needed, data comes from the
// selected site.
type GitPanel struct {
	site *forge.Site
}

// NewGitPanel creates a new GitPanel.
func NewGitPanel(site *forge.Site) GitPanel {
	return GitPanel{site: site}
}

// SetSite replaces the displayed site.
func (p GitPanel) SetSite(site *forge.Site) GitPanel {
	p.site = site
	return p
}

// Update handles messages. GitPanel is display-only.
func (p GitPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	return p, nil
}

// View renders the git info panel.
func (p GitPanel) View(width, height int, focused bool) string {
	style := theme.InactiveBorderStyle
	titleColor := theme.ColorSubtle
	if focused {
		style = theme.ActiveBorderStyle
		titleColor = theme.ColorPrimary
	}

	innerWidth := width - 2
	innerHeight := height - 2
	if innerWidth < 0 {
		innerWidth = 0
	}
	if innerHeight < 0 {
		innerHeight = 0
	}

	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(titleColor).
		Render(" Git ")

	var lines []string

	if p.site == nil {
		lines = append(lines, theme.NormalItemStyle.Render("No site selected"))
	} else {
		site := p.site

		provider := site.RepositoryProvider
		if provider == "" {
			provider = "-"
		}
		lines = append(lines, renderInfoKV("Provider", provider, innerWidth))

		repo := site.Repository
		if repo == "" {
			repo = "-"
		}
		lines = append(lines, renderInfoKV("Repository", repo, innerWidth))

		branch := site.RepositoryBranch
		if branch == "" {
			branch = "-"
		}
		lines = append(lines, renderInfoKV("Branch", branch, innerWidth))

		repoStatus := site.RepositoryStatus
		if repoStatus == "" {
			repoStatus = "-"
		}
		lines = append(lines, renderStatusKV("Status", repoStatus, innerWidth))

		// Additional relevant info.
		lines = append(lines, "")
		lines = append(lines, renderInfoKV("Quick Deploy", boolToOnOff(site.QuickDeploy), innerWidth))
		if site.DeploymentURL != "" {
			lines = append(lines, renderInfoKV("Deploy URL", site.DeploymentURL, innerWidth))
		}
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

// HelpBindings returns the key hints for the git panel.
func (p GitPanel) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "1-9", Desc: "sections"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
