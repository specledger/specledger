# AI Chatbot Template

## Overview

This template provides a LangGraph-based AI chatbot with support for multiple integrations (Slack, Discord, Telegram, Web), RAG capabilities, and tool use. It's designed for building conversational AI applications with memory and multi-turn interactions.

## Technology Stack

- **Language**: Python 3.10+
- **Framework**: LangGraph / LangChain
- **LLM Providers**: OpenAI, Anthropic, or local models
- **Vector Store**: ChromaDB, Pinecone, or Weaviate
- **Integrations**: Slack, Discord, Telegram, Web

## Directory Structure

```
.
├── src/
│   ├── agents/
│   │   └── base.py         # Agent definitions
│   ├── chains/             # LangChain chains
│   ├── tools/              # Custom tools for agents
│   ├── prompts/            # Prompt templates
│   ├── memory/             # Conversation memory
│   ├── vectorstore/        # RAG and embeddings
│   ├── middleware/         # Pre/post processing
│   ├── integrations/
│   │   ├── slack/          # Slack bot integration
│   │   ├── discord/        # Discord bot integration
│   │   ├── telegram/       # Telegram bot integration
│   │   └── web/            # Web API integration
│   │       └── app.py
│   └── utils/              # Helper utilities
├── data/
│   └── documents/          # Documents for RAG
├── configs/
│   └── example.env         # Environment configuration
├── tests/                  # Test files
├── langgraph.json          # LangGraph configuration
├── requirements.txt
└── pyproject.toml
```

## Development Commands

### Setup Environment
```bash
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
```

### Run Web Interface
```bash
python -m src.integrations.web.app
```

### Run with LangGraph Studio
```bash
langgraph dev
```

### Testing
```bash
pytest tests/
```

## Agent Development Guidelines

### Creating Agents
1. Define agent in `src/agents/`
2. Add tools in `src/tools/`
3. Configure prompts in `src/prompts/`

### Adding Tools
```python
from langchain.tools import tool

@tool
def search_documents(query: str) -> str:
    """Search through indexed documents."""
    # Implementation
```

### RAG Setup
1. Place documents in `data/documents/`
2. Index with vectorstore in `src/vectorstore/`
3. Connect to agent retrieval

### Memory Management
- Use conversation buffer for short-term
- Vector store for long-term memory
- Clear sessions appropriately

## Code Guidelines

- Keep prompts versioned and testable
- Test tools independently
- Log all LLM interactions for debugging
- Handle API rate limits and errors

<!-- MANUAL ADDITIONS START -->
<!-- MANUAL ADDITIONS END -->
