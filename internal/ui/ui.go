package ui

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	mutedStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	boldStyle    = lipgloss.NewStyle().Bold(true)
)

var colorEnabled = sync.OnceValue(func() bool {
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
})

func ColorEnabled() bool {
	return colorEnabled()
}

func Success(msg string) {
	if ColorEnabled() {
		fmt.Println(successStyle.Render("✓ " + msg))
		return
	}
	fmt.Println("✓ " + msg)
}

func Warn(msg string) {
	if ColorEnabled() {
		fmt.Println(warnStyle.Render("! " + msg))
		return
	}
	fmt.Println("! " + msg)
}

func Error(msg string) {
	if ColorEnabled() {
		fmt.Fprintln(os.Stderr, errorStyle.Render("x "+msg))
		return
	}
	fmt.Fprintln(os.Stderr, "x "+msg)
}

func Info(msg string) {
	if ColorEnabled() {
		fmt.Println(mutedStyle.Render("- " + msg))
		return
	}
	fmt.Println("- " + msg)
}

func Bold(msg string) string {
	if ColorEnabled() {
		return boldStyle.Render(msg)
	}
	return msg
}

type Spinner struct {
	s *spinner.Spinner
}

func NewSpinner(msg string) *Spinner {
	s := spinner.New(spinner.CharSets[14], 80*time.Millisecond, spinner.WithWriter(os.Stderr))
	s.Suffix = "  " + msg
	_ = s.Color("cyan")
	s.Start()
	return &Spinner{s: s}
}

func (sp *Spinner) Stop() {
	sp.s.Stop()
}
