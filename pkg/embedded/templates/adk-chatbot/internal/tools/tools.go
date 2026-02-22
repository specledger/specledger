//go:build ignore

package tools

import (
	"fmt"
	"time"

	adktool "google.golang.org/adk/tool"
	"google.golang.org/adk/tool/functiontool"
)

// TimeArgs is the input for the current_time tool.
type TimeArgs struct {
	Timezone string `json:"timezone"` // IANA timezone (e.g. "America/New_York"), defaults to UTC
}

// TimeResult is the output of the current_time tool.
type TimeResult struct {
	Time     string `json:"time"`     // formatted current time
	Timezone string `json:"timezone"` // timezone used
}

func currentTime(ctx adktool.Context, input TimeArgs) (TimeResult, error) {
	tz := input.Timezone
	if tz == "" {
		tz = "UTC"
	}
	loc, err := time.LoadLocation(tz)
	if err != nil {
		return TimeResult{}, fmt.Errorf("invalid timezone %q: %w", tz, err)
	}
	now := time.Now().In(loc)
	return TimeResult{
		Time:     now.Format(time.RFC1123),
		Timezone: tz,
	}, nil
}

// CalculateArgs is the input for the calculator tool.
type CalculateArgs struct {
	Expression string `json:"expression"` // a simple math expression to evaluate (e.g. "2 + 2")
}

// CalculateResult is the output of the calculator tool.
type CalculateResult struct {
	Result string `json:"result"` // the evaluated result
}

func calculate(_ adktool.Context, input CalculateArgs) (CalculateResult, error) {
	return CalculateResult{
		Result: fmt.Sprintf("Expression: %s (implement a math parser for production use)", input.Expression),
	}, nil
}

// NewChatTools returns the default set of tools for the chatbot.
func NewChatTools() ([]adktool.Tool, error) {
	timeTool, err := functiontool.New(functiontool.Config{
		Name:        "current_time",
		Description: "Get the current time in a specified timezone.",
	}, currentTime)
	if err != nil {
		return nil, fmt.Errorf("creating time tool: %w", err)
	}

	calcTool, err := functiontool.New(functiontool.Config{
		Name:        "calculator",
		Description: "Evaluate a simple math expression.",
	}, calculate)
	if err != nil {
		return nil, fmt.Errorf("creating calculator tool: %w", err)
	}

	return []adktool.Tool{timeTool, calcTool}, nil
}
