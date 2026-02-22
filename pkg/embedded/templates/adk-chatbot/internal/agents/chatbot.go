//go:build ignore

package agents

import (
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model"
	"google.golang.org/adk/tool"
)

// NewChatbotAgent creates a chatbot agent with the given model and tools.
func NewChatbotAgent(m model.LLM, chatTools []tool.Tool) (*llmagent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "chatbot",
		Model:       m,
		Description: "A helpful AI chatbot powered by Google ADK.",
		Instruction: `You are a helpful, friendly AI assistant.
Answer user questions clearly and concisely.
Use available tools when they can help answer the user's question.
If you don't know something, say so honestly.`,
		Tools: chatTools,
	})
}

// NewRAGAgent creates an agent with retrieval-augmented generation capabilities.
func NewRAGAgent(m model.LLM, ragTools []tool.Tool) (*llmagent.Agent, error) {
	return llmagent.New(llmagent.Config{
		Name:        "rag_agent",
		Model:       m,
		Description: "An agent that retrieves and synthesizes information from documents.",
		Instruction: `You are a knowledge assistant. When asked a question:
1. Use the search tool to find relevant information.
2. Synthesize the results into a clear, accurate answer.
3. Cite sources when possible.`,
		Tools: ragTools,
	})
}
