//go:build mage
// +build mage

package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
	"github.com/magefile/mage/sh"
)

const (
	MODULE_NAME = "hmerritt/reactenv" // go.mod module name
	LOG_LEVEL   = 4                   // 5 = debug, 4 = info, 3 = warn, 2 = error
)

// ----------------------------------------------------------------------------
// Runtime
// ----------------------------------------------------------------------------

// Runs a single cmd command, and streams the output to stdout
func RunStream(args []string, dir string, addPadding bool) error {
	cmd := exec.Command(args[0], args[1:]...)

	if dir != "" {
		cmd.Dir = dir
	}

	if addPadding {
		fmt.Println("")
	}

	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	err := cmd.Start()

	if err != nil {
		return err
	}

	pipeHandler := func(pipe io.ReadCloser, textColor color.Attribute) {
		scanner := bufio.NewScanner(pipe)
		scanner.Split(bufio.ScanLines)
		for scanner.Scan() {
			color.Set(textColor)
			fmt.Println(scanner.Text())
			color.Unset()
		}
	}

	go func() {
		pipeHandler(stdout, color.Reset)
	}()

	go func() {
		pipeHandler(stderr, color.Reset)
	}()

	cmd.Wait()

	if addPadding {
		fmt.Println("")
	}

	exitCode := cmd.ProcessState.ExitCode()

	if exitCode != 0 {
		return fmt.Errorf("command exited with code %d", exitCode)
	}

	return nil
}

// Runs multiple cmd commands one-by-one
func RunSync(commands [][]string) error {
	for _, cmd := range commands {
		if len(cmd) == 0 {
			continue
		}

		if err := sh.Run(cmd[0], cmd[1:]...); err != nil {
			return err
		}
	}

	return nil
}

// Runs multiple commands in parallel
func RunParallel(commands [][]string) error {
	var wg sync.WaitGroup
	var errCatch error = nil

	// Launch a goroutine for each command.
	for _, cmd := range commands {
		if len(cmd) == 0 {
			continue
		}

		wg.Add(1)

		go func(cmd []string) {
			defer wg.Done()
			if err := sh.Run(cmd[0], cmd[1:]...); err != nil {
				errCatch = err
			}
		}(cmd)
	}

	// Wait for all the goroutines to finish.
	wg.Wait()

	// If any of the commands failed, return the first error.
	if errCatch != nil {
		return errCatch
	}

	return nil
}

// ----------------------------------------------------------------------------
// CLI
// ----------------------------------------------------------------------------

type Logger struct {
	// The logging level the logger should log at. This is typically (and defaults
	// to) `Info`, which allows Info(), Warn(), Error() and Fatal() to be logged.
	Level uint32

	// Name of the function Logger was initiated from.
	FnInitName string

	// Timestamp of Logger initiation.
	InitTimestamp time.Time

	// Timestamp of the most recent log. Used to calculate and show the time in
	// milliseconds since last log.
	PrevTimestamp time.Time
}

func NewLogger() *Logger {
	// Function name
	pc, _, _, _ := runtime.Caller(1)
	funcName := runtime.FuncForPC(pc).Name()
	funcName = funcName[strings.LastIndex(funcName, ".")+1:] // Removes package name

	return &Logger{
		Level:         LOG_LEVEL,
		FnInitName:    funcName,
		InitTimestamp: time.Now(),
		PrevTimestamp: time.Now(),
	}
}

func (l *Logger) log(level uint32, a ...interface{}) {
	if l.Level < level {
		return
	}

	currentTime := time.Now()
	formattedTime := currentTime.Format("2006-01-02 15:04:05")
	toLog := fmt.Sprintf("%s (%s) +%7s => ", formattedTime, l.FnInitName, DurationSince(l.PrevTimestamp))

	messages := make([]interface{}, 0)
	messages = append(messages, toLog)
	messages = append(messages, a...)
	fmt.Println(messages...)
	l.PrevTimestamp = currentTime
}

func (l *Logger) SetLevel(level uint32) {
	l.Level = level
}
func (l *Logger) Error(messages ...interface{}) error {
	color.Set(color.FgRed)
	defer color.Unset()
	l.log(2, messages...)
	return errors.New(strings.Trim(strings.Join(strings.Fields(fmt.Sprint(messages)), " "), "[]"))
}
func (l *Logger) Warn(messages ...interface{}) {
	color.Set(color.FgYellow)
	defer color.Unset()
	l.log(3, messages...)
}
func (l *Logger) Info(messages ...interface{}) {
	l.log(4, messages...)
}
func (l *Logger) Debug(messages ...interface{}) {
	l.log(5, messages...)
}
func (l *Logger) End() {
	color.Set(color.FgCyan)
	defer color.Unset()
	l.log(4, fmt.Sprintf("took %s", DurationSince(l.InitTimestamp)))
}

func DurationSince(since time.Time) string {
	ms := time.Since(since).Milliseconds()

	if ms < 1000 {
		return fmt.Sprintf("%.2fms", float64(ms))
	}

	if ms < 60000 {
		s := float64(ms) / 1000
		return fmt.Sprintf("%.2fs", s)
	}

	m := float64(ms) / 60000
	return fmt.Sprintf("%.2fm", m)
}

// ----------------------------------------------------------------------------
// MISC
// ----------------------------------------------------------------------------

// Checks if an executable exists in PATH
func ExecExists(e string) bool {
	_, err := exec.LookPath(e)
	return err == nil
}

// Get ENV, or use fallback value
func GetEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
