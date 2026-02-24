package panels

import (
	"context"
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/key"
	lipgloss "charm.land/lipgloss/v2"

	"github.com/hinke/phorge/internal/forge"
	"github.com/hinke/phorge/internal/tui/theme"
)

// --- Messages ---

// CertsLoadedMsg is sent when the certificate list has been fetched.
type CertsLoadedMsg struct {
	Certificates []forge.Certificate
}

// CertCreatedMsg is sent when a Let's Encrypt certificate has been created.
type CertCreatedMsg struct {
	Certificate *forge.Certificate
}

// CertActivatedMsg is sent when a certificate has been activated.
type CertActivatedMsg struct{}

// CertDeletedMsg is sent when a certificate has been deleted.
type CertDeletedMsg struct{}

// SSLPanel shows the SSL certificates for a site with CRUD actions.
type SSLPanel struct {
	client   *forge.Client
	serverID int64
	siteID   int64

	certificates []forge.Certificate
	cursor       int
	loading      bool

	// Keybindings
	up       key.Binding
	down     key.Binding
	create   key.Binding
	activate key.Binding
	del      key.Binding
	home     key.Binding
	end      key.Binding
}

// NewSSLPanel creates a new SSLPanel.
func NewSSLPanel(client *forge.Client, serverID, siteID int64) SSLPanel {
	return SSLPanel{
		client:   client,
		serverID: serverID,
		siteID:   siteID,
		loading:  true,
		up: key.NewBinding(
			key.WithKeys("k", "up"),
			key.WithHelp("k/up", "up"),
		),
		down: key.NewBinding(
			key.WithKeys("j", "down"),
			key.WithHelp("j/down", "down"),
		),
		create: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "create"),
		),
		activate: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "activate"),
		),
		del: key.NewBinding(
			key.WithKeys("x"),
			key.WithHelp("x", "delete"),
		),
		home: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "top"),
		),
		end: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "bottom"),
		),
	}
}

// LoadCerts returns a tea.Cmd that fetches the certificate list.
func (p SSLPanel) LoadCerts() tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		certs, err := client.Certificates.List(context.Background(), serverID, siteID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return CertsLoadedMsg{Certificates: certs}
	}
}

// CreateLetsEncrypt returns a tea.Cmd that creates a Let's Encrypt certificate.
func (p SSLPanel) CreateLetsEncrypt(domains []string) tea.Cmd {
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	return func() tea.Msg {
		cert, err := client.Certificates.CreateLetsEncrypt(context.Background(), serverID, siteID, domains)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return CertCreatedMsg{Certificate: cert}
	}
}

// ActivateCert returns a tea.Cmd that activates the currently selected certificate.
func (p SSLPanel) ActivateCert() tea.Cmd {
	if len(p.certificates) == 0 || p.cursor >= len(p.certificates) {
		return nil
	}
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	certID := p.certificates[p.cursor].ID
	return func() tea.Msg {
		err := client.Certificates.Activate(context.Background(), serverID, siteID, certID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return CertActivatedMsg{}
	}
}

// DeleteCert returns a tea.Cmd that deletes the currently selected certificate.
func (p SSLPanel) DeleteCert() tea.Cmd {
	if len(p.certificates) == 0 || p.cursor >= len(p.certificates) {
		return nil
	}
	client := p.client
	serverID := p.serverID
	siteID := p.siteID
	certID := p.certificates[p.cursor].ID
	return func() tea.Msg {
		err := client.Certificates.Delete(context.Background(), serverID, siteID, certID)
		if err != nil {
			return PanelErrMsg{Err: err}
		}
		return CertDeletedMsg{}
	}
}

// SelectedCert returns the currently selected certificate, or nil.
func (p SSLPanel) SelectedCert() *forge.Certificate {
	if len(p.certificates) == 0 || p.cursor >= len(p.certificates) {
		return nil
	}
	cert := p.certificates[p.cursor]
	return &cert
}

// Update handles messages for the SSL panel.
func (p SSLPanel) Update(msg tea.Msg) (Panel, tea.Cmd) {
	switch msg := msg.(type) {
	case CertsLoadedMsg:
		p.certificates = msg.Certificates
		p.loading = false
		p.cursor = 0
		return p, nil

	case tea.KeyPressMsg:
		return p.handleKey(msg)
	}

	return p, nil
}

func (p SSLPanel) handleKey(msg tea.KeyPressMsg) (Panel, tea.Cmd) {
	switch {
	case key.Matches(msg, p.down):
		if len(p.certificates) > 0 {
			p.cursor = min(p.cursor+1, len(p.certificates)-1)
		}
		return p, nil

	case key.Matches(msg, p.up):
		if len(p.certificates) > 0 {
			p.cursor = max(p.cursor-1, 0)
		}
		return p, nil

	case key.Matches(msg, p.home):
		p.cursor = 0
		return p, nil

	case key.Matches(msg, p.end):
		if len(p.certificates) > 0 {
			p.cursor = len(p.certificates) - 1
		}
		return p, nil

	// 'c', 'a', 'x' are handled by the app layer.
	}

	return p, nil
}

// View renders the SSL panel.
func (p SSLPanel) View(width, height int, focused bool) string {
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
		Render(" SSL Certificates ")

	content := p.renderList(innerWidth, innerHeight)

	return style.
		Width(innerWidth).
		Height(innerHeight).
		Render(title + "\n" + content)
}

func (p SSLPanel) renderList(width, height int) string {
	var lines []string

	if p.loading && len(p.certificates) == 0 {
		lines = append(lines, theme.LoadingStyle.Render("Loading certificates..."))
	} else if len(p.certificates) == 0 {
		lines = append(lines, theme.NormalItemStyle.Render("No certificates found"))
	} else {
		visibleHeight := height - 1
		if visibleHeight < 1 {
			visibleHeight = 1
		}
		startIdx := 0
		if p.cursor >= visibleHeight {
			startIdx = p.cursor - visibleHeight + 1
		}

		for i := startIdx; i < len(p.certificates) && len(lines) < visibleHeight; i++ {
			cert := p.certificates[i]
			line := p.renderCertLine(cert, i, width)
			lines = append(lines, line)
		}
	}

	for len(lines) < height {
		lines = append(lines, "")
	}

	return strings.Join(lines, "\n")
}

func (p SSLPanel) renderCertLine(cert forge.Certificate, idx, maxWidth int) string {
	// Active indicator.
	var activeStr string
	if cert.Active {
		activeStr = lipgloss.NewStyle().Foreground(theme.ColorSecondary).Render("*")
	} else {
		activeStr = lipgloss.NewStyle().Foreground(theme.ColorSubtle).Render(" ")
	}

	icon := statusIcon(cert.Status)

	domain := cert.Domain
	if domain == "" {
		domain = "-"
	}
	certType := cert.Type
	if certType == "" {
		certType = "unknown"
	}
	statusStr := fmt.Sprintf(" [%s]", cert.Status)

	// Leave room for: cursor(2) + active(2) + icon(2) + type(~12) + status(~14) + spacing(6)
	overhead := 38
	domainWidth := maxWidth - overhead
	if domainWidth < 10 {
		domainWidth = 10
	}
	domain = truncatePlain(domain, domainWidth)

	typeStr := fmt.Sprintf("%-10s", truncatePlain(certType, 10))

	if idx == p.cursor {
		line := theme.CursorStyle.Render("> ") +
			activeStr + " " +
			icon + " " +
			theme.SelectedItemStyle.Render(domain) +
			"  " + theme.NormalItemStyle.Render(typeStr) +
			"  " + theme.NormalItemStyle.Render(statusStr)
		return theme.Truncate(line, maxWidth)
	}

	line := "  " +
		activeStr + " " +
		icon + " " +
		theme.NormalItemStyle.Render(domain) +
		"  " + theme.NormalItemStyle.Render(typeStr) +
		"  " + theme.NormalItemStyle.Render(statusStr)
	return theme.Truncate(line, maxWidth)
}

// HelpBindings returns the key hints for the SSL panel.
func (p SSLPanel) HelpBindings() []HelpBinding {
	return []HelpBinding{
		{Key: "j/k", Desc: "navigate"},
		{Key: "c", Desc: "create LE cert"},
		{Key: "a", Desc: "activate"},
		{Key: "x", Desc: "delete"},
		{Key: "g/G", Desc: "top/bottom"},
		{Key: "esc", Desc: "back"},
		{Key: "tab", Desc: "switch panel"},
		{Key: "q", Desc: "quit"},
	}
}
