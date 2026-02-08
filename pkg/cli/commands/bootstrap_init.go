package commands

import (
	"fmt"
	"os/exec"

	"specledger/pkg/cli/metadata"
)

// installAndInitFrameworks installs framework tools via mise and runs their init commands
func installAndInitFrameworks(projectPath string, framework metadata.FrameworkChoice) error {
	if framework == metadata.FrameworkNone {
		return nil
	}

	fmt.Println("\nInstalling framework tools...")

	// Run mise install in project directory
	cmd := exec.Command("mise", "install")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to install tools via mise: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		fmt.Println("You can manually install later with: cd", projectPath, "&& mise install")
		return nil // Don't fail bootstrap if mise install fails
	}

	fmt.Println("✓ Framework tools installed")

	// Initialize frameworks based on selection
	switch framework {
	case metadata.FrameworkSpecKit:
		return initSpecKit(projectPath)
	case metadata.FrameworkOpenSpec:
		return initOpenSpec(projectPath)
	case metadata.FrameworkBoth:
		if err := initSpecKit(projectPath); err != nil {
			return err
		}
		return initOpenSpec(projectPath)
	}

	return nil
}

// initSpecKit runs specify init in the project directory
func initSpecKit(projectPath string) error {
	fmt.Println("\nInitializing Spec Kit...")

	// Use --here flag to initialize in current directory
	cmd := exec.Command("specify", "init", "--here")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to initialize Spec Kit: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		fmt.Println("You can manually initialize later with: cd", projectPath, "&& specify init --here")
		return nil // Don't fail bootstrap
	}

	fmt.Println("✓ Spec Kit initialized")
	return nil
}

// initOpenSpec runs openspec init in the project directory
func initOpenSpec(projectPath string) error {
	fmt.Println("\nInitializing OpenSpec...")

	cmd := exec.Command("openspec", "init")
	cmd.Dir = projectPath
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Printf("⚠️  Warning: Failed to initialize OpenSpec: %v\n", err)
		fmt.Printf("Output: %s\n", string(output))
		fmt.Println("You can manually initialize later with: cd", projectPath, "&& openspec init")
		return nil // Don't fail bootstrap
	}

	fmt.Println("✓ OpenSpec initialized")
	return nil
}
