//go:build ignore

package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"google.golang.org/genai"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"

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

	sessionService := session.InMemoryService()
	r, err := runner.New(runner.Config{
		AppName:        "chatbot",
		Agent:          a,
		SessionService: sessionService,
	})
	if err != nil {
		log.Fatalf("Failed to create runner: %v", err)
	}

	sessResp, err := sessionService.Create(ctx, &session.CreateRequest{
		AppName: "chatbot",
		UserID:  "user1",
	})
	if err != nil {
		log.Fatalf("Failed to create session: %v", err)
	}

	msg := &genai.Content{
		Role:  "user",
		Parts: []*genai.Part{genai.NewPartFromText("Hello! What can you help me with?")},
	}

	for event, err := range r.Run(ctx, "user1", sessResp.Session.ID(), msg, agent.RunConfig{}) {
		if err != nil {
			log.Fatalf("Error during run: %v", err)
		}
		if event.IsFinalResponse() {
			for _, part := range event.Content.Parts {
				if part.Text != "" {
					fmt.Println("Agent:", part.Text)
				}
			}
		}
	}
}
