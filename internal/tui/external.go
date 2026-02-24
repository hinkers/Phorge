package tui

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
)

// dbReadyMsg is sent after successfully fetching and parsing .env database credentials.
type dbReadyMsg struct {
	host       string
	port       string
	database   string
	username   string
	password   string
	connection string // e.g. "mysql", "pgsql"
	sshUser    string
	sshHost    string
	sshPort    int
}

// sshCmd returns a tea.Cmd that suspends the TUI and opens an SSH session
// to the currently selected server. If a site with a directory is selected,
// the SSH session will cd into that directory.
func (m App) sshCmd() tea.Cmd {
	if m.selectedSrv == nil {
		return nil
	}

	user := m.config.SSHUserFor(m.selectedSrv.Name)
	args := []string{fmt.Sprintf("%s@%s", user, m.selectedSrv.IPAddress)}

	// Custom SSH port.
	if m.selectedSrv.SSHPort != 0 && m.selectedSrv.SSHPort != 22 {
		args = append([]string{"-p", fmt.Sprintf("%d", m.selectedSrv.SSHPort)}, args...)
	}

	// If a site is selected with a directory, cd into it on the remote.
	if m.selectedSite != nil && m.selectedSite.Directory != "" {
		args = append(args, "-t", fmt.Sprintf("cd %s && exec $SHELL -l", m.selectedSite.Directory))
	}

	c := exec.Command("ssh", args...)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return externalExitMsg{err}
	})
}

// sftpCmd returns a tea.Cmd that suspends the TUI and opens termscp (SCP/SFTP)
// to the currently selected server. The path defaults to "/" but uses the
// site directory if a site is selected.
func (m App) sftpCmd() tea.Cmd {
	if m.selectedSrv == nil {
		return nil
	}

	user := m.config.SSHUserFor(m.selectedSrv.Name)
	port := m.selectedSrv.SSHPort
	if port == 0 {
		port = 22
	}

	remotePath := "/"
	if m.selectedSite != nil && m.selectedSite.Directory != "" {
		remotePath = m.selectedSite.Directory
	}

	target := fmt.Sprintf("scp://%s@%s:%d%s", user, m.selectedSrv.IPAddress, port, remotePath)
	c := exec.Command("termscp", target)
	return tea.ExecProcess(c, func(err error) tea.Msg {
		return externalExitMsg{err}
	})
}

// databaseCmd returns a tea.Cmd that fetches the .env file for the selected
// site, parses DB credentials, and sends a dbReadyMsg so the app can set up
// the SSH tunnel and launch lazysql.
func (m App) databaseCmd() tea.Cmd {
	if m.selectedSrv == nil || m.selectedSite == nil {
		return nil
	}

	client := m.forge
	srv := m.selectedSrv
	site := m.selectedSite
	user := m.config.SSHUserFor(srv.Name)

	return func() tea.Msg {
		// Fetch the .env file from the Forge API.
		envContent, err := client.Environment.Get(context.Background(), srv.ID, site.ID)
		if err != nil {
			return errMsg{fmt.Errorf("failed to fetch .env: %w", err)}
		}

		// Parse the DB credentials from the .env content.
		dbCreds := parseEnvVars(envContent)
		if dbCreds["DB_HOST"] == "" {
			return errMsg{fmt.Errorf("DB_HOST not found in .env")}
		}

		return dbReadyMsg{
			host:       dbCreds["DB_HOST"],
			port:       dbCreds["DB_PORT"],
			database:   dbCreds["DB_DATABASE"],
			username:   dbCreds["DB_USERNAME"],
			password:   dbCreds["DB_PASSWORD"],
			connection: dbCreds["DB_CONNECTION"],
			sshUser:    user,
			sshHost:    srv.IPAddress,
			sshPort:    srv.SSHPort,
		}
	}
}

// handleDBReady sets up the SSH tunnel and launches lazysql for the database
// connection described in msg. It returns the updated model and tea.Cmd.
func (m App) handleDBReady(msg dbReadyMsg) (App, tea.Cmd) {
	// Find a free local port for the SSH tunnel.
	localPort, err := findFreePort()
	if err != nil {
		m.toast = fmt.Sprintf("Failed to find free port: %v", err)
		m.toastIsErr = true
		return m, m.clearToastAfter(5 * time.Second)
	}

	// Determine the remote DB port (default based on connection type).
	dbPort := msg.port
	if dbPort == "" {
		switch msg.connection {
		case "pgsql":
			dbPort = "5432"
		default:
			dbPort = "3306" // mysql is the default
		}
	}

	// Build the SSH tunnel command.
	sshPort := msg.sshPort
	if sshPort == 0 {
		sshPort = 22
	}

	tunnelSpec := fmt.Sprintf("%d:%s:%s", localPort, msg.host, dbPort)
	tunnelArgs := []string{
		"-L", tunnelSpec,
		"-N", // no remote command
		"-o", "StrictHostKeyChecking=no",
		"-o", "ExitOnForwardFailure=yes",
	}
	if sshPort != 22 {
		tunnelArgs = append(tunnelArgs, "-p", fmt.Sprintf("%d", sshPort))
	}
	tunnelArgs = append(tunnelArgs, fmt.Sprintf("%s@%s", msg.sshUser, msg.sshHost))

	tunnel := exec.Command("ssh", tunnelArgs...)
	tunnel.Stdout = nil
	tunnel.Stderr = nil

	if err := tunnel.Start(); err != nil {
		m.toast = fmt.Sprintf("Failed to start SSH tunnel: %v", err)
		m.toastIsErr = true
		return m, m.clearToastAfter(5 * time.Second)
	}

	// Store the tunnel process for cleanup.
	m.tunnelProc = tunnel.Process

	// Wait briefly for the tunnel to establish.
	time.Sleep(time.Second)

	// Build the lazysql connection string.
	// lazysql accepts a DSN-style connection string.
	lazysqlArgs := buildLazysqlArgs(msg, localPort)
	lazysqlCmd := exec.Command("lazysql", lazysqlArgs...)

	// Store reference for cleanup in the callback.
	tunnelProc := tunnel.Process

	return m, tea.ExecProcess(lazysqlCmd, func(err error) tea.Msg {
		// Always kill the tunnel when lazysql exits.
		if tunnelProc != nil {
			_ = tunnelProc.Kill()
		}
		return externalExitMsg{err}
	})
}

// parseEnvVars parses a .env file content into a map of key-value pairs.
// It handles comments, empty lines, and quoted values.
func parseEnvVars(content string) map[string]string {
	vars := make(map[string]string)
	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		i := strings.IndexByte(line, '=')
		if i < 0 {
			continue
		}
		key := strings.TrimSpace(line[:i])
		value := strings.TrimSpace(line[i+1:])
		// Remove surrounding quotes (single or double).
		value = strings.Trim(value, `"'`)
		vars[key] = value
	}
	return vars
}

// findFreePort asks the OS for an available TCP port.
func findFreePort() (int, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

// buildLazysqlArgs constructs the command-line arguments for lazysql
// based on the database connection type and credentials.
func buildLazysqlArgs(msg dbReadyMsg, localPort int) []string {
	// lazysql uses environment variables for connection.
	// We set them via the command's environment and pass the connection string.
	switch msg.connection {
	case "pgsql":
		// For PostgreSQL: lazysql postgres://user:pass@host:port/dbname
		dsn := fmt.Sprintf("postgres://%s:%s@127.0.0.1:%d/%s?sslmode=disable",
			msg.username, msg.password, localPort, msg.database)
		return []string{dsn}
	default:
		// For MySQL (default): lazysql mysql://user:pass@host:port/dbname
		dsn := fmt.Sprintf("mysql://%s:%s@127.0.0.1:%d/%s",
			msg.username, msg.password, localPort, msg.database)
		return []string{dsn}
	}
}

// cleanupTunnel kills the SSH tunnel process if it's running.
func (m *App) cleanupTunnel() {
	if m.tunnelProc != nil {
		_ = m.tunnelProc.Kill()
		// Reap the process to avoid zombies.
		_, _ = m.tunnelProc.Wait()
		m.tunnelProc = nil
	}
}

// setLazysqlEnv sets environment variables on the lazysql command
// for database credentials that may not be in the DSN.
func setLazysqlEnv(cmd *exec.Cmd, msg dbReadyMsg, localPort int) {
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("DB_HOST=127.0.0.1"),
		fmt.Sprintf("DB_PORT=%d", localPort),
		fmt.Sprintf("DB_USER=%s", msg.username),
		fmt.Sprintf("DB_PASSWORD=%s", msg.password),
		fmt.Sprintf("DB_NAME=%s", msg.database),
	)
}
