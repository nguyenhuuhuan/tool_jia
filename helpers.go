package main

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/rivo/tview"
)

func OpenBrowser(app *tview.Application, url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)

	cmdObj := exec.Command(cmd, args...)

	app.Suspend(func() {
		cmdObj.Start()
	})
	return nil
}

func GenerateBranchName(issue Issue) string {
	// Sanitize summary: lowercase, replace spaces with hyphens, remove special chars
	sanitizedSummary := strings.ToLower(issue.Fields.Summary)
	sanitizedSummary = strings.ReplaceAll(sanitizedSummary, " ", "-")
	reg := strings.NewReplacer(
		"&", "and",
		"@", "at",
		"!", "",
		"\"", "",
		"'", "",
		"?", "",
		",", "",
		".", "",
		":", "",
		";", "",
		"(", "",
		")", "",
		"[", "",
		"]", "",
		"{", "",
		"}", "",
	)
	sanitizedSummary = reg.Replace(sanitizedSummary)
	// Limit length to avoid overly long branch names
	if len(sanitizedSummary) > 60 {
		sanitizedSummary = sanitizedSummary[:100]
	}
	return fmt.Sprintf("feature/%s-%s", issue.Key, sanitizedSummary)
}
