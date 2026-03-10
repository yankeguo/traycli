package main

import (
	"os"
	"os/exec"
	"runtime"
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
	cc, err := ReadConfig(cfg)
	if err != nil {
		dialog.Message("Failed to read config: %v", err).Title("traycli Error").Error()
		os.Exit(1)
	}
	if cc == nil {
		dialog.Message("config.json not found. Creating template at %s", cfg.ConfigPath).Title("traycli Error").Error()
		if err := WriteEmptyConfig(cfg); err != nil {
			dialog.Message("Failed to create config: %v", err).Title("traycli Error").Error()
			os.Exit(1)
		}
		openFile(cfg.ConfigPath)
		os.Exit(1)
	}
	if len(cc.Cmd) == 0 {
		dialog.Message("config.json has empty cmd. Please edit %s and add a command.", cfg.ConfigPath).Title("traycli Error").Error()
		os.Exit(1)
	}

	runner = NewRunner(cfg, cc)
	systray.Run(onReady, onExit)
}

func openFile(path string) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", path)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", "", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	cmd.Run()
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
