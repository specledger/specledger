package context

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

type TechnicalContext struct {
	Language         string
	PrimaryDeps      string
	Storage          string
	Testing          string
	TargetPlatform   string
	ProjectType      string
	PerformanceGoals string
	Constraints      string
	Scale            string
}

var (
	techContextStart = regexp.MustCompile(`^##\s+Technical\s+Context\s*$`)
	fieldPattern     = regexp.MustCompile(`^\*\*([^*]+)\*\*:\s*(.+)$`)
)

func ParseTechnicalContext(planPath string) (*TechnicalContext, error) {
	file, err := os.Open(planPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("plan.md not found: %s", planPath)
		}
		return nil, fmt.Errorf("failed to open plan.md: %w", err)
	}
	defer file.Close()

	ctx := &TechnicalContext{}

	scanner := bufio.NewScanner(file)
	inTechContext := false
	foundAny := false

	for scanner.Scan() {
		line := scanner.Text()

		if techContextStart.MatchString(line) {
			inTechContext = true
			continue
		}

		if inTechContext {
			if strings.HasPrefix(line, "## ") && !techContextStart.MatchString(line) {
				break
			}

			matches := fieldPattern.FindStringSubmatch(line)
			if len(matches) == 3 {
				fieldName := strings.TrimSpace(matches[1])
				fieldValue := strings.TrimSpace(matches[2])

				foundAny = true

				switch fieldName {
				case "Language/Version":
					ctx.Language = fieldValue
				case "Primary Dependencies":
					ctx.PrimaryDeps = fieldValue
				case "Storage":
					ctx.Storage = fieldValue
				case "Testing":
					ctx.Testing = fieldValue
				case "Target Platform":
					ctx.TargetPlatform = fieldValue
				case "Project Type":
					ctx.ProjectType = fieldValue
				case "Performance Goals":
					ctx.PerformanceGoals = fieldValue
				case "Constraints":
					ctx.Constraints = fieldValue
				case "Scale/Scope":
					ctx.Scale = fieldValue
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read plan.md: %w", err)
	}

	if !foundAny {
		return nil, fmt.Errorf("no Technical Context fields found in plan.md")
	}

	return ctx, nil
}

func (ctx *TechnicalContext) String() string {
	var parts []string

	if ctx.Language != "" {
		parts = append(parts, fmt.Sprintf("Language: %s", ctx.Language))
	}
	if ctx.PrimaryDeps != "" {
		parts = append(parts, fmt.Sprintf("Dependencies: %s", ctx.PrimaryDeps))
	}
	if ctx.Storage != "" {
		parts = append(parts, fmt.Sprintf("Storage: %s", ctx.Storage))
	}
	if ctx.Testing != "" {
		parts = append(parts, fmt.Sprintf("Testing: %s", ctx.Testing))
	}
	if ctx.TargetPlatform != "" {
		parts = append(parts, fmt.Sprintf("Platform: %s", ctx.TargetPlatform))
	}
	if ctx.ProjectType != "" {
		parts = append(parts, fmt.Sprintf("Project Type: %s", ctx.ProjectType))
	}

	return strings.Join(parts, "\n")
}

func (ctx *TechnicalContext) ToMarkdown() string {
	var lines []string

	if ctx.Language != "" {
		lines = append(lines, fmt.Sprintf("- **Language/Version**: %s", ctx.Language))
	}
	if ctx.PrimaryDeps != "" {
		lines = append(lines, fmt.Sprintf("- **Primary Dependencies**: %s", ctx.PrimaryDeps))
	}
	if ctx.Storage != "" {
		lines = append(lines, fmt.Sprintf("- **Storage**: %s", ctx.Storage))
	}
	if ctx.Testing != "" {
		lines = append(lines, fmt.Sprintf("- **Testing**: %s", ctx.Testing))
	}
	if ctx.TargetPlatform != "" {
		lines = append(lines, fmt.Sprintf("- **Target Platform**: %s", ctx.TargetPlatform))
	}
	if ctx.ProjectType != "" {
		lines = append(lines, fmt.Sprintf("- **Project Type**: %s", ctx.ProjectType))
	}
	if ctx.PerformanceGoals != "" {
		lines = append(lines, fmt.Sprintf("- **Performance Goals**: %s", ctx.PerformanceGoals))
	}
	if ctx.Constraints != "" {
		lines = append(lines, fmt.Sprintf("- **Constraints**: %s", ctx.Constraints))
	}
	if ctx.Scale != "" {
		lines = append(lines, fmt.Sprintf("- **Scale/Scope**: %s", ctx.Scale))
	}

	return strings.Join(lines, "\n")
}
