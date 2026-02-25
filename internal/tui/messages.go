package tui

import "github.com/hinkers/Phorge/internal/forge"

// serversLoadedMsg is sent when the server list has been fetched from the API.
type serversLoadedMsg struct {
	servers []forge.Server
}

// errMsg is sent when an API call or other operation fails.
type errMsg struct {
	err error
}

// deployResultMsg is sent when a deploy operation completes.
type deployResultMsg struct {
	err error
}

// rebootResultMsg is sent when a server reboot completes.
type rebootResultMsg struct {
	err error
}

// toastMsg is sent to display a temporary notification.
type toastMsg struct {
	message string
	isError bool
}

// externalExitMsg is sent after returning from an external process (ssh, sftp, etc.).
type externalExitMsg struct {
	err error
}

// clearToastMsg is sent to dismiss the toast notification.
type clearToastMsg struct{}

// setDefaultMsg is sent after toggling the default server/site in .phorge.
type setDefaultMsg struct {
	serverName string // empty means cleared
	siteName   string // empty means cleared
	err        error
}

// pollOutputTickMsg is sent by the output polling timer to trigger a refresh.
type pollOutputTickMsg struct{}

// pollOutputResultMsg carries the result of a polled output fetch.
type pollOutputResultMsg struct {
	output   string
	finished bool
}
