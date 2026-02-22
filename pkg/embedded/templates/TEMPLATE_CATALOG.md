# SpecLedger Project Template Catalog

> Machine-readable template catalog for the SpecLedger project creation UI.
> The frontend reads this file to render the template selection screen during `sl new`.

---

## Metadata

| Field        | Value           |
| ------------ | --------------- |
| version      | 1.1.0           |
| total        | 8               |
| default      | general-purpose |
| last_updated | 2026-02-21      |

---

## Categories

| Category        | Description                                      | Templates                                    |
| --------------- | ------------------------------------------------ | -------------------------------------------- |
| General         | Starter projects and CLI tools                   | general-purpose                              |
| Web             | Full-stack and frontend/backend applications     | full-stack                                   |
| Data & Workflow | ETL pipelines, streaming, and workflow engines   | batch-data, realtime-workflow, realtime-data |
| AI & ML         | Machine learning, chatbots, and LLM applications | ml-image, ai-chatbot, adk-chatbot            |

---

## Templates

### general-purpose

| Field           | Value                                                                                                                                                                                                                 |
| --------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Name**        | General Purpose                                                                                                                                                                                                       |
| **ID**          | `general-purpose`                                                                                                                                                                                                     |
| **Category**    | General                                                                                                                                                                                                               |
| **Language**    | Go                                                                                                                                                                                                                    |
| **Tags**        | `Go`, `CLI`, `SpecLedger`                                                                                                                                                                                             |
| **Default**     | Yes                                                                                                                                                                                                                   |
| **Description** | Go CLI/library project with SpecLedger workflows and Claude Code integration. The baseline template that every project starts from -- includes the SpecLedger playbook, agent configuration, and development tooling. |

#### When to use

- Starting a new Go CLI tool or library
- Projects that primarily need the SpecLedger specification workflow
- When no other template fits your use case

#### What's included

```
.claude/               # Claude Code agent configuration
.specledger/           # SpecLedger metadata and memory
  memory/
    constitution.md    # Project constitution template
AGENTS.md              # Agent development guidelines
.gitattributes         # Git LFS and diff config
mise.toml              # Tool version management (Go, Node, etc.)
scripts/               # Development helper scripts
```

#### Example project

**Project**: `sl-migrate` -- a CLI tool that migrates legacy YAML configs to the new schema.

```bash
sl new --name sl-migrate --template general-purpose
cd sl-migrate
go mod init github.com/myorg/sl-migrate
# Write your CLI logic in cmd/ and internal/
go run . migrate --input legacy.yaml --output v2.yaml
```

---

### full-stack

| Field           | Value                                                                                                                                                                                     |
| --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Name**        | Full-Stack Application                                                                                                                                                                    |
| **ID**          | `full-stack`                                                                                                                                                                              |
| **Category**    | Web                                                                                                                                                                                       |
| **Language**    | Go + TypeScript                                                                                                                                                                           |
| **Tags**        | `Go`, `React`, `TypeScript`, `Vite`                                                                                                                                                       |
| **Default**     | No                                                                                                                                                                                        |
| **Description** | Go backend with TypeScript/React frontend, Docker Compose for local PostgreSQL, and clean architecture separation. Backend uses Gin for HTTP routing; frontend uses Vite for fast builds. |

#### When to use

- Building a web application with a separate API and SPA
- Projects that need Go backend performance with a React frontend
- Teams that want clear backend/frontend separation from day one

#### What's included

```
backend/
  cmd/server/main.go       # HTTP server entrypoint
  internal/
    handlers/              # HTTP request handlers
    services/              # Business logic layer
    models/                # Data models and DTOs
    middleware/             # Auth, logging, CORS middleware
    database/              # Database connection and queries
  go.mod

frontend/
  src/
    components/            # Reusable React components
    pages/                 # Route-level page components
    hooks/                 # Custom React hooks
    services/              # API client (fetch/axios wrappers)
  public/                  # Static assets
  package.json

docker-compose.yml         # PostgreSQL + backend + frontend
README.md
AGENTS.md
```

#### Example project

**Project**: `task-board` -- a Kanban-style task management app with drag-and-drop.

```bash
sl new --name task-board --template full-stack
cd task-board

# Start infrastructure
docker-compose up -d

# Backend (Go API server)
cd backend && go run ./cmd/server
# Runs on http://localhost:8080

# Frontend (React SPA)
cd frontend && npm install && npm run dev
# Runs on http://localhost:5173
```

Key endpoints the template scaffolds:

| Method | Path              | Purpose           |
| ------ | ----------------- | ----------------- |
| GET    | /api/health       | Health check      |
| GET    | /api/v1/items     | List resources    |
| POST   | /api/v1/items     | Create a resource |
| PUT    | /api/v1/items/:id | Update a resource |
| DELETE | /api/v1/items/:id | Delete a resource |

---

### batch-data

| Field           | Value                                                                                                                                                                                                         |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Name**        | Batch Data Processing                                                                                                                                                                                         |
| **ID**          | `batch-data`                                                                                                                                                                                                  |
| **Category**    | Data & Workflow                                                                                                                                                                                               |
| **Language**    | Go                                                                                                                                                                                                            |
| **Tags**        | `Go`, `Temporal`, `ETL`                                                                                                                                                                                       |
| **Default**     | No                                                                                                                                                                                                            |
| **Description** | Temporal-based ETL pipeline with Go workers and a clean Extract-Transform-Load pattern. Includes Docker Compose for Temporal and PostgreSQL infrastructure, plus pre-wired workflow and activity definitions. |

#### When to use

- Scheduled data ingestion or migration jobs
- ETL pipelines that pull from APIs, transform, and load into databases
- Long-running batch jobs that need retry and failure handling

#### What's included

```
workflows/
  workflow.go              # Temporal workflow definitions (BatchProcessWorkflow)
  activity.go              # Activity definitions (Extract, Transform, Load)

cmd/
  worker/main.go           # Temporal worker process
  starter/                 # CLI to trigger workflow executions

internal/
  extractors/              # Data extraction logic (APIs, files, DBs)
  transformers/            # Data cleaning and transformation
  loaders/                 # Destination writers (DB, S3, API)

config/                    # YAML/env configuration
tests/                     # Unit and integration tests
docker-compose.yml         # Temporal Server + PostgreSQL
go.mod
README.md
AGENTS.md
```

#### Example project

**Project**: `sales-etl` -- nightly pipeline that extracts CRM data, normalizes it, and loads into a data warehouse.

```bash
sl new --name sales-etl --template batch-data
cd sales-etl

# Start Temporal and PostgreSQL
docker-compose up -d

# Run the worker (listens for workflow tasks)
go run ./cmd/worker

# Trigger a batch run (in another terminal)
go run ./cmd/starter
```

Workflow execution flow:

```
┌──────────┐     ┌─────────────┐     ┌──────────┐
│ Extract  │────▶│  Transform  │────▶│   Load   │
│ (API/DB) │     │  (Clean)    │     │  (DW)    │
└──────────┘     └─────────────┘     └──────────┘
     ▲                                     │
     └─── Temporal retries on failure ─────┘
```

---

### realtime-workflow

| Field           | Value                                                                                                                                                                                                               |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Name**        | Real-Time Workflow                                                                                                                                                                                                  |
| **ID**          | `realtime-workflow`                                                                                                                                                                                                 |
| **Category**    | Data & Workflow                                                                                                                                                                                                     |
| **Language**    | Go                                                                                                                                                                                                                  |
| **Tags**        | `Go`, `Temporal`, `Workflows`                                                                                                                                                                                       |
| **Default**     | No                                                                                                                                                                                                                  |
| **Description** | Temporal workflow orchestration for event-driven applications, long-running business processes, and distributed transactions. Supports saga patterns, human-in-the-loop signals, and scheduled recurring workflows. |

#### When to use

- Event-driven business processes (order fulfillment, onboarding flows)
- Long-running workflows that span hours, days, or weeks
- Distributed transactions that need compensation/rollback (saga pattern)
- Human-in-the-loop approval workflows

#### What's included

```
cmd/
  worker/main.go                 # Temporal worker process
  starter/                       # Workflow trigger CLI

internal/
  workflows/
    workflow.go                  # Workflow definitions (signals, queries, timers)
  activities/
    activities.go                # Activity implementations (I/O, API calls)
  models/                        # Data models and state types
  config/                        # Configuration management

tests/                           # Workflow and activity tests
docker-compose.yml               # Temporal Server + PostgreSQL + Elasticsearch
go.mod
README.md
AGENTS.md
```

#### Example project

**Project**: `order-saga` -- an order fulfillment system that coordinates payment, inventory, and shipping with compensations.

```bash
sl new --name order-saga --template realtime-workflow
cd order-saga

# Start Temporal infrastructure
docker-compose up -d

# Run the worker
go run ./cmd/worker

# Trigger an order workflow
go run ./cmd/starter --order-id ORD-12345
```

Workflow pattern:

```
┌─────────┐    ┌───────────┐    ┌──────────┐    ┌──────────┐
│ Reserve  │───▶│  Charge   │───▶│  Ship    │───▶│ Complete │
│ Stock    │    │  Payment  │    │  Order   │    │  Order   │
└─────────┘    └───────────┘    └──────────┘    └──────────┘
     │               │               │
     ▼               ▼               ▼
 Compensate:     Compensate:     Compensate:
 Release Stock   Refund Payment  Cancel Shipment
```

---

### realtime-data

| Field           | Value                                                                                                                                                                                                                 |
| --------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Name**        | Real-Time Data Pipeline                                                                                                                                                                                               |
| **ID**          | `realtime-data`                                                                                                                                                                                                       |
| **Category**    | Data & Workflow                                                                                                                                                                                                       |
| **Language**    | Go                                                                                                                                                                                                                    |
| **Tags**        | `Go`, `Kafka`, `Streaming`                                                                                                                                                                                            |
| **Default**     | No                                                                                                                                                                                                                    |
| **Description** | Kafka-based streaming pipeline with separate producer, consumer, and processor services. Uses segmentio/kafka-go for high-performance Go Kafka clients. Includes Docker Compose for the full Kafka + Zookeeper stack. |

#### When to use

- Real-time event ingestion and processing
- Building event-driven microservice architectures
- Data streaming between systems (CDC, log aggregation, metrics)
- High-throughput message processing with consumer groups

#### What's included

```
cmd/
  producer/main.go              # Kafka producer service
  consumer/main.go              # Kafka consumer service
  processor/                    # Stream processor service

internal/
  kafka/
    producer.go                 # Producer implementation (async, batched)
    consumer.go                 # Consumer group implementation
    config.go                   # Broker and topic configuration
  handlers/                     # Message event handlers
  models/                       # Message schemas and DTOs
  processors/                   # Stream processing logic

configs/                        # Environment configuration
deployments/
  docker-compose.yml            # Kafka + Zookeeper + Schema Registry
tests/                          # Unit and integration tests
go.mod
README.md
AGENTS.md
```

#### Example project

**Project**: `clickstream` -- real-time clickstream analytics that ingests user events, enriches them, and feeds dashboards.

```bash
sl new --name clickstream --template realtime-data
cd clickstream

# Start Kafka infrastructure
docker-compose -f deployments/docker-compose.yml up -d

# Run the producer (ingests events)
go run ./cmd/producer

# Run the consumer (processes events)
go run ./cmd/consumer
```

Data flow:

```
┌──────────┐     ┌───────────────┐     ┌───────────┐
│ Producer │────▶│  Kafka Topic  │────▶│ Consumer  │
│ (events) │     │ (partitioned) │     │ (group)   │
└──────────┘     └───────────────┘     └───────────┘
                                            │
                                       ┌────▼─────┐
                                       │Processor │
                                       │(enrich)  │
                                       └──────────┘
```

---

### ml-image

| Field           | Value                                                                                                                                                                                                                                           |
| --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Name**        | ML Image Processing                                                                                                                                                                                                                             |
| **ID**          | `ml-image`                                                                                                                                                                                                                                      |
| **Category**    | AI & ML                                                                                                                                                                                                                                         |
| **Language**    | Python                                                                                                                                                                                                                                          |
| **Tags**        | `Python`, `PyTorch`, `ML`, `Computer Vision`                                                                                                                                                                                                    |
| **Default**     | No                                                                                                                                                                                                                                              |
| **Description** | PyTorch-based machine learning project for computer vision tasks including classification, object detection, and segmentation. Follows the standard data science project layout with clear separation of data, models, training, and inference. |

#### When to use

- Image classification, object detection, or segmentation tasks
- Training custom computer vision models
- Projects that need experiment tracking and reproducible pipelines
- Research prototyping with Jupyter notebooks

#### What's included

```
src/
  data/
    dataset.py                 # Dataset loading and augmentation
    preprocessing.py           # Image preprocessing pipelines
    augmentation.py            # Data augmentation strategies
  models/
    cnn_model.py               # Model architecture definitions (ResNet, etc.)
  training/
    train.py                   # Training loop with checkpointing
    evaluate.py                # Evaluation metrics (accuracy, mAP, IoU)
  inference/
    predict.py                 # Inference pipeline for serving
  utils/                       # Visualization, logging utilities

data/
  raw/                         # Original unprocessed images
  processed/                   # Cleaned and resized images
  interim/                     # Intermediate processing outputs

models/checkpoints/            # Saved model weights (.pt files)
notebooks/                     # Jupyter notebooks for exploration
configs/                       # Training hyperparameter configs
tests/                         # Unit tests
requirements.txt
pyproject.toml
README.md
AGENTS.md
```

#### Example project

**Project**: `defect-detector` -- a manufacturing defect detection system that classifies product images.

```bash
sl new --name defect-detector --template ml-image
cd defect-detector

# Setup environment
python -m venv venv && source venv/bin/activate
pip install -r requirements.txt

# Place training images
cp -r /path/to/images data/raw/

# Train the model
python -m src.training.train --config configs/train.yaml

# Evaluate
python -m src.training.evaluate --model models/checkpoints/best.pt

# Run inference on new images
python -m src.inference.predict --input new_image.jpg
```

Training pipeline:

```
┌──────────┐     ┌──────────────┐     ┌──────────┐     ┌──────────┐
│ Raw Data │────▶│ Preprocess   │────▶│  Train   │────▶│ Evaluate │
│ (images) │     │ + Augment    │     │  Model   │     │ Metrics  │
└──────────┘     └──────────────┘     └──────────┘     └──────────┘
                                           │
                                      ┌────▼──────┐
                                      │Checkpoint │
                                      │ (.pt)     │
                                      └───────────┘
```

---

### ai-chatbot

| Field           | Value                                                                                                                                                                        |
| --------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Name**        | AI Chatbot                                                                                                                                                                   |
| **ID**          | `ai-chatbot`                                                                                                                                                                 |
| **Category**    | AI & ML                                                                                                                                                                      |
| **Language**    | Python                                                                                                                                                                       |
| **Tags**        | `Python`, `LangChain`, `LLM`, `RAG`                                                                                                                                          |
| **Default**     | No                                                                                                                                                                           |
| **Description** | LangChain/LangGraph multi-platform chatbot with RAG support, tool use, and conversation memory. Includes adapters for Slack, Discord, Telegram, and a FastAPI web interface. |

#### When to use

- Building a conversational AI chatbot with RAG (retrieval-augmented generation)
- Multi-platform bots (Slack, Discord, Telegram) from a single codebase
- Projects using the LangChain/LangGraph Python ecosystem
- Chatbots that need custom tools and long-term memory

#### What's included

```
src/
  agents/
    base.py                    # LangGraph agent with StateGraph
  chains/                      # LangChain chain compositions
  tools/                       # Custom tools (@tool decorated functions)
  prompts/                     # Versioned prompt templates
  memory/                      # Conversation buffer and vector memory
  vectorstore/                 # RAG embedding and retrieval
  middleware/                  # Message pre/post processing
  integrations/
    slack/                     # Slack Bot adapter
    discord/                   # Discord Bot adapter
    telegram/                  # Telegram Bot adapter
    web/
      app.py                   # FastAPI web interface

data/documents/                # Documents for RAG indexing
configs/
  example.env                  # API keys and platform tokens
tests/                         # Test files
langgraph.json                 # LangGraph Studio configuration
requirements.txt
pyproject.toml
README.md
AGENTS.md
```

#### Example project

**Project**: `support-bot` -- a customer support chatbot that answers questions from your docs and escalates to humans.

```bash
sl new --name support-bot --template ai-chatbot
cd support-bot

# Setup environment
python -m venv venv && source venv/bin/activate
pip install -r requirements.txt

# Configure API keys
cp configs/example.env .env
# Edit .env with your OPENAI_API_KEY, SLACK_BOT_TOKEN, etc.

# Add your documents for RAG
cp /path/to/docs/*.pdf data/documents/

# Run the web interface
python -m src.integrations.web.app
# API available at http://localhost:8000

# Or use LangGraph Studio
langgraph dev
```

Architecture:

```
                    ┌──────────────────┐
                    │   LangGraph      │
                    │   Agent Engine   │
                    └────────┬─────────┘
                             │
            ┌────────────────┼────────────────┐
            ▼                ▼                 ▼
      ┌──────────┐    ┌──────────┐     ┌───────────┐
      │  Tools   │    │  Memory  │     │ RAG Store │
      │(search,  │    │(buffer + │     │(vectorDB) │
      │ actions) │    │ vector)  │     │           │
      └──────────┘    └──────────┘     └───────────┘
            │
  ┌─────────┼─────────┬──────────┐
  ▼         ▼         ▼          ▼
┌──────┐ ┌───────┐ ┌────────┐ ┌─────┐
│Slack │ │Discord│ │Telegram│ │ Web │
└──────┘ └───────┘ └────────┘ └─────┘
```

---

### adk-chatbot

| Field           | Value                                                                                                                                                                                                                                                             |
| --------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Name**        | AI Chatbot (Google ADK)                                                                                                                                                                                                                                           |
| **ID**          | `adk-chatbot`                                                                                                                                                                                                                                                     |
| **Category**    | AI & ML                                                                                                                                                                                                                                                           |
| **Language**    | Go                                                                                                                                                                                                                                                                |
| **Tags**        | `Go`, `Google ADK`, `Gemini`, `LLM`                                                                                                                                                                                                                               |
| **Default**     | No                                                                                                                                                                                                                                                                |
| **Description** | Go-based AI chatbot using Google's Agent Development Kit (ADK) with Gemini models. Includes custom function tools, session management, and multi-agent workflow support. The ADK launcher provides a CLI console, web UI, and A2A protocol server out of the box. |

#### When to use

- Building AI agents and chatbots in Go (not Python)
- Projects using Google Gemini models
- Applications that need the ADK launcher (CLI + web UI + A2A for free)
- Multi-agent orchestration with sequential, parallel, or loop workflows
- Teams that prefer Go's type safety and concurrency for AI workloads

#### What's included

```
cmd/
  chatbot/main.go              # ADK full launcher (CLI + web + A2A)
  server/main.go               # Programmatic runner example

internal/
  agents/
    chatbot.go                 # Agent definitions (Chatbot, RAG agent)
  tools/
    tools.go                   # Custom function tools (current_time, calculator)
  middleware/                   # Request/response middleware

configs/
  example.env.sample           # Google API key and server config
tests/                         # Test files
go.mod
README.md
AGENTS.md
```

#### Example project

**Project**: `travel-agent` -- an AI travel assistant that searches flights, checks weather, and plans itineraries using multi-agent workflows.

```bash
sl new --name travel-agent --template adk-chatbot
cd travel-agent

# Set up
export GOOGLE_API_KEY=your-google-api-key
go mod tidy

# Run interactive CLI
go run ./cmd/chatbot

# Run web UI (opens at http://localhost:8080)
go run ./cmd/chatbot web api webui

# Run the programmatic example
go run ./cmd/server
```

Architecture:

```
┌─────────────────────────────────────────┐
│            ADK Launcher                 │
│  ┌───────┐  ┌───────┐  ┌───────────┐   │
│  │  CLI  │  │Web UI │  │ A2A Proto │   │
│  └───┬───┘  └───┬───┘  └─────┬─────┘   │
│      └──────────┼─────────────┘         │
│                 ▼                        │
│          ┌─────────────┐                │
│          │   Runner    │                │
│          │  (session)  │                │
│          └──────┬──────┘                │
│                 ▼                        │
│          ┌─────────────┐                │
│          │  LLM Agent  │                │
│          │  (Gemini)   │                │
│          └──────┬──────┘                │
│      ┌──────────┼──────────┐            │
│      ▼          ▼          ▼            │
│ ┌─────────┐┌────────┐┌──────────┐      │
│ │ Custom  ││ Google ││   MCP    │      │
│ │ Tools   ││ Search ││ Toolset  │      │
│ └─────────┘└────────┘└──────────┘      │
└─────────────────────────────────────────┘
```

Adding a custom tool:

```go
type FlightArgs struct {
    From string `json:"from"` // departure city
    To   string `json:"to"`   // arrival city
    Date string `json:"date"` // travel date (YYYY-MM-DD)
}
type FlightResult struct {
    Flights []string `json:"flights"` // available flights
}

flightTool, _ := functiontool.New(functiontool.Config{
    Name:        "search_flights",
    Description: "Search for available flights between cities.",
}, func(ctx tool.Context, in FlightArgs) (FlightResult, error) {
    // call your flights API here
    return FlightResult{Flights: []string{"AA100 09:00", "UA200 14:30"}}, nil
})
```

---

## Quick Comparison

| Template          | Language | Category | Key Framework    | Docker | Use Case                                             |
| ----------------- | -------- | -------- | ---------------- | ------ | ---------------------------------------------------- |
| general-purpose   | Go       | General  | Cobra CLI        | No     | CLI tools, libraries, SpecLedger projects            |
| full-stack        | Go + TS  | Web      | Gin + React/Vite | Yes    | Web apps with API + SPA                              |
| batch-data        | Go       | Data     | Temporal         | Yes    | ETL pipelines, batch jobs                            |
| realtime-workflow | Go       | Data     | Temporal         | Yes    | Sagas, approvals, long-running processes             |
| realtime-data     | Go       | Data     | Kafka            | Yes    | Event streaming, CDC, log aggregation                |
| ml-image          | Python   | AI & ML  | PyTorch          | No     | Image classification, detection, segmentation        |
| ai-chatbot        | Python   | AI & ML  | LangChain        | No     | RAG chatbots, multi-platform bots                    |
| adk-chatbot       | Go       | AI & ML  | Google ADK       | No     | Go AI agents, Gemini chatbots, multi-agent workflows |

---

## Frontend Rendering Notes

- **Default template**: Highlight `general-purpose` with a "Recommended" badge
- **Tags**: Render each tag from the `Tags` field as a pill/chip component
- **Categories**: Group templates by category for easier scanning
- **Diagram blocks**: The ASCII diagrams in each template section (inside triple-backtick blocks without a language) can be rendered as monospace pre-formatted text
- **Example project name**: Use the bold **Project** name as the subtitle in the template card
- **Quick Comparison table**: Can be used for a comparison/filter view
