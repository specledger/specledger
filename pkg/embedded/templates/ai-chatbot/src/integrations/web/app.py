"""FastAPI web interface for the chatbot."""

from fastapi import FastAPI, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel

from src.agents.base import BaseAgent

app = FastAPI(title="AI Chatbot API")

# Add CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Initialize agent
agent = BaseAgent()


class ChatRequest(BaseModel):
    """Request model for chat endpoint."""

    message: str
    context: str = ""


class ChatResponse(BaseModel):
    """Response model for chat endpoint."""

    response: str


@app.get("/health")
async def health():
    """Health check endpoint."""
    return {"status": "ok"}


@app.post("/chat", response_model=ChatResponse)
async def chat(request: ChatRequest):
    """Chat with the AI agent."""
    try:
        response = await agent.chat(request.message, request.context)
        return ChatResponse(response=response)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


if __name__ == "__main__":
    import uvicorn

    uvicorn.run(app, host="0.0.0.0", port=8000)
