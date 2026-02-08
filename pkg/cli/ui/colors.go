package ui

import (
	"fmt"
)

// Color functions for common use cases
func Success(text string) string {
	return "\033[32m" + text + "\033[0m"
}

func SuccessBold(text string) string {
	return "\033[1m\033[32m" + text + "\033[0m"
}

func Error(text string) string {
	return "\033[31m" + text + "\033[0m"
}

func ErrorBold(text string) string {
	return "\033[1m\033[31m" + text + "\033[0m"
}

func Warning(text string) string {
	return "\033[33m" + text + "\033[0m"
}

func WarningBold(text string) string {
	return "\033[1m\033[33m" + text + "\033[0m"
}

func Info(text string) string {
	return "\033[36m" + text + "\033[0m"
}

func InfoBold(text string) string {
	return "\033[1m\033[36m" + text + "\033[0m"
}

func Primary(text string) string {
	return "\033[34m" + text + "\033[0m"
}

func PrimaryBold(text string) string {
	return "\033[1m\033[34m" + text + "\033[0m"
}

func MagentaText(text string) string {
	return "\033[35m" + text + "\033[0m"
}

func MagentaBold(text string) string {
	return "\033[1m\033[35m" + text + "\033[0m"
}

func Dim(text string) string {
	return "\033[90m" + text + "\033[0m"
}

func Bold(text string) string {
	return "\033[1m" + text + "\033[0m"
}

func Cyan(text string) string {
	return "\033[36m" + text + "\033[0m"
}

func Green(text string) string {
	return "\033[32m" + text + "\033[0m"
}

func Red(text string) string {
	return "\033[31m" + text + "\033[0m"
}

func Yellow(text string) string {
	return "\033[33m" + text + "\033[0m"
}

func Blue(text string) string {
	return "\033[34m" + text + "\033[0m"
}

func Gray(text string) string {
	return "\033[90m" + text + "\033[0m"
}

// ANSI color codes for use in fmt.Println and string concatenation
const (
	Reset    = "\033[0m"
	BoldCode = "\033[1m"
	DimCode  = "\033[90m"
)

// Checkmark and X mark with colors
func Checkmark() string {
	return "\033[32m✓\033[0m"
}

func Crossmark() string {
	return "\033[31m✗\033[0m"
}

func InfoIcon() string {
	return "\033[34mℹ\033[0m"
}

func WarningIcon() string {
	return "\033[33m⚠\033[0m"
}

func Rocket() string {
	return "\033[34m🚀\033[0m"
}

// Box drawing characters for UI
func BoxTop(width int) string {
	return "┌" + repeat("─", width-2) + "┐"
}

func BoxBottom(width int) string {
	return "└" + repeat("─", width-2) + "┘"
}

func repeat(s string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += s
	}
	return result
}

// visibleLen returns the visible length of a string, excluding ANSI color codes
func visibleLen(s string) int {
	len := 0
	inEscape := false
	for _, c := range s {
		if c == '\x1b' {
			inEscape = true
		} else if inEscape && c == 'm' {
			inEscape = false
		} else if !inEscape {
			len++
		}
	}
	return len
}

// PrintHeader prints a header without box
func PrintHeader(title, subtitle string, width int) {
	fmt.Println()
	if subtitle != "" {
		fmt.Println(Bold(Blue(title)) + " " + Gray(subtitle))
	} else {
		fmt.Println(Bold(Blue(title)))
	}
	fmt.Println()
}

// PrintSection prints a section header
func PrintSection(title string) {
	fmt.Println()
	fmt.Println(Bold(Blue(title)))
	fmt.Println(Cyan(repeat("─", len(title))))
	fmt.Println()
}

// PrintSuccess prints a success message
func PrintSuccess(message string) {
	fmt.Println()
	fmt.Println(SuccessBold("✓ " + message))
	fmt.Println()
}

// PrintError prints an error message
func PrintError(message string) {
	fmt.Println()
	fmt.Println(ErrorBold("✗ " + message))
	fmt.Println()
}

// PrintWarning prints a warning message
func PrintWarning(message string) {
	fmt.Println()
	fmt.Println(WarningBold("⚠ " + message))
	fmt.Println()
}

// PrintBox prints text in a colored box
func PrintBox(text string, colorFunc func(string) string, width int) {
	lines := []string{text}
	if width == 0 {
		width = len(text) + 4
		if width < 40 {
			width = 40
		}
	}

	fmt.Println()
	fmt.Println(colorFunc("┌" + repeat("─", width-2) + "┐"))

	for _, line := range lines {
		padding := width - len(line) - 4
		if padding < 0 {
			padding = 0
		}
		leftPad := padding / 2
		rightPad := padding - leftPad
		fmt.Println(colorFunc("│") + repeat(" ", leftPad) + line + repeat(" ", rightPad) + colorFunc("│"))
	}

	fmt.Println(colorFunc("└" + repeat("─", width-2) + "┘"))
	fmt.Println()
}
