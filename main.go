package main

import (
	"os"
	"strconv"
	"time"

	"github.com/getlantern/systray"
	"github.com/sqweek/dialog"
)

var (
	runner     *Runner
	statusItem *systray.MenuItem
)

func main() {
	cfg, err := LoadConfig()
	if err != nil {
		dialog.Message("Failed to load config: %v", err).Title("traycli Error").Error()
		os.Exit(1)
	}
	command, err := ReadCommand(cfg)
	if err != nil {
		dialog.Message("Failed to read command: %v", err).Title("traycli Error").Error()
		os.Exit(1)
	}
	if command == "" {
		dialog.Message("command.txt not found or empty. Please create %s with the command to run.", cfg.CommandPath).Title("traycli Error").Error()
		os.Exit(1)
	}

	runner = NewRunner(cfg, command)
	systray.Run(onReady, onExit)
}

func onReady() {
	systray.SetTitle("traycli")
	systray.SetTooltip("traycli")

	statusItem = systray.AddMenuItem("", "")
	statusItem.Disable()
	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Quit traycli")

	go runner.Run()
	go updateStatus()

	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()
}

func onExit() {
	if runner != nil {
		runner.Stop()
	}
}

func updateStatus() {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if runner == nil || statusItem == nil {
			return
		}
		uptime := runner.Uptime()
		restarts := runner.Restarts()
		status := formatDuration(uptime)
		status += " | Restarts: " + strconv.FormatUint(restarts, 10)
		statusItem.SetTitle(status)
	}
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return "Uptime: " + d.Round(time.Second).String()
	}
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m >= 60 {
		h := m / 60
		m = m % 60
		return "Uptime: " + strconv.FormatUint(uint64(h), 10) + "h " + strconv.FormatUint(uint64(m), 10) + "m " + strconv.FormatUint(uint64(s), 10) + "s"
	}
	return "Uptime: " + strconv.FormatUint(uint64(m), 10) + "m " + strconv.FormatUint(uint64(s), 10) + "s"
}
