# AI Chatbot Template

LangChain/LangGraph-based multi-platform AI chatbot.

## Structure

```
src/
├── agents/               # AI agent definitions
├── tools/                # Custom tools for agents
├── prompts/              # Prompt templates
├── chains/               # LangChain chains
├── memory/               # Conversation memory
├── middleware/           # Message middleware
├── integrations/
│   ├── slack/            # Slack bot adapter
│   ├── discord/          # Discord bot adapter
│   ├── telegram/         # Telegram bot adapter
│   └── web/              # Web chat interface
├── vectorstore/          # Vector database integration
└── utils/                # Utility functions

data/documents/           # RAG source documents
tests/                    # Test files
configs/                  # Configuration files
```

## Technologies

- **Framework**: LangChain + LangGraph
- **LLM**: OpenAI / Anthropic
- **Vector DB**: Pinecone / Weaviate
- **Platforms**: Slack, Discord, Telegram SDKs

## Getting Started

1. Install: `pip install -r requirements.txt`
2. Configure: Copy `configs/example.env` to `.env`
3. Run: `python -m src.integrations.web`
