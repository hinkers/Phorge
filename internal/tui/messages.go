package tui

import "github.com/hinke/phorge/internal/forge"

// serversLoadedMsg is sent when the server list has been fetched from the API.
type serversLoadedMsg struct {
	servers []forge.Server
}

// sitesLoadedMsg is sent when the site list for a server has been fetched.
type sitesLoadedMsg struct {
	sites []forge.Site
}

// errMsg is sent when an API call or other operation fails.
type errMsg struct {
	err error
}

// serverSelectedMsg is sent when a server is selected from the list.
type serverSelectedMsg struct {
	server *forge.Server
}

// siteSelectedMsg is sent when a site is selected from the context list.
type siteSelectedMsg struct {
	site *forge.Site
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
