//go:build ignore

package main

import (
	"context"
	"log"
	"os"

	"google.golang.org/genai"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/model/gemini"

	"example.com/project/internal/tools"
)

func main() {
	ctx := context.Background()

	model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})
	if err != nil {
		log.Fatalf("Failed to create Gemini model: %v", err)
	}

	chatTools, err := tools.NewChatTools()
	if err != nil {
		log.Fatalf("Failed to create tools: %v", err)
	}

	a, err := llmagent.New(llmagent.Config{
		Name:        "chatbot",
		Model:       model,
		Description: "A helpful AI chatbot powered by Google ADK.",
		Instruction: `You are a helpful, friendly AI assistant.
Answer user questions clearly and concisely.
Use available tools when they can help answer the user's question.`,
		Tools: chatTools,
	})
	if err != nil {
		log.Fatalf("Failed to create agent: %v", err)
	}

	l := full.NewLauncher()
	if err = l.Execute(ctx, &launcher.Config{
		AgentLoader: agent.NewSingleLoader(a),
	}, os.Args[1:]); err != nil {
		log.Fatalf("Run failed: %v\n\n%s", err, l.CommandLineSyntax())
	}
}
