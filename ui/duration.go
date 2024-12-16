package ui

import (
	"fmt"
	"time"

	"github.com/mitchellh/cli"
)

// Record duration of a task/command
type Duration struct {
	ui    *Ui
	start time.Time
}

func InitDuration(ui *Ui) *Duration {
	return &Duration{
		ui:    ui,
		start: time.Now(),
	}
}

func (d *Duration) In(textColor cli.UiColor, text string) {
	if text == "" {
		d.ui.Output(fmt.Sprintf("took %s", time.Since(d.start)))
	} else {
		d.ui.Output(fmt.Sprintf("%s in %s", d.ui.Colorize(text, textColor), time.Since(d.start)))
	}
}

func (d *Duration) Since() time.Duration {
	return time.Since(d.start)
}
