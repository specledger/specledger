package launcher

import (
	"fmt"
	"os/exec"

	"github.com/specledger/specledger/internal/agent"
)

type Agent struct {
	Definition     agent.Agent
	InstallCommand string
}

func NewAgentFromDefinition(def agent.Agent) *Agent {
	return &Agent{
		Definition:     def,
		InstallCommand: def.InstallCommand,
	}
}

func (a *Agent) CheckInstalled() error {
	if a.Definition.Command == "" {
		return fmt.Errorf("agent has no command configured")
	}
	_, err := exec.LookPath(a.Definition.Command)
	if err != nil {
		return fmt.Errorf("'%s' not found. Install: %s", a.Definition.Command, a.InstallCommand)
	}
	return nil
}

func (a *Agent) Name() string {
	return a.Definition.Name
}

func (a *Agent) Command() string {
	return a.Definition.Command
}

func NewLauncherForAgent(ag agent.Agent, dir string) *AgentLauncher {
	return NewAgentLauncher(AgentOption{
		Name:        ag.Name,
		Command:     ag.Command,
		Description: "",
	}, dir)
}
