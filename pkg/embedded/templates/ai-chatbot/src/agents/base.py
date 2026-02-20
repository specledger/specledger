"""Base agent implementation using LangGraph."""

from typing import TypedDict, Annotated, Sequence
from operator import add

from langchain_core.messages import BaseMessage, HumanMessage, AIMessage
from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, END


class AgentState(TypedDict):
    """State for the agent graph."""

    messages: Annotated[Sequence[BaseMessage], add]
    context: str


class BaseAgent:
    """Base agent class using LangGraph."""

    def __init__(self, model_name: str = "gpt-4"):
        """
        Initialize the agent.

        Args:
            model_name: Name of the LLM model to use
        """
        self.llm = ChatOpenAI(model=model_name)
        self.graph = self._build_graph()

    def _build_graph(self) -> StateGraph:
        """Build the agent graph."""
        graph = StateGraph(AgentState)

        # Add nodes
        graph.add_node("respond", self._respond)

        # Set entry point
        graph.set_entry_point("respond")

        # Add edges
        graph.add_edge("respond", END)

        return graph.compile()

    async def _respond(self, state: AgentState) -> dict:
        """Generate a response to the user message."""
        messages = state["messages"]
        response = await self.llm.ainvoke(messages)
        return {"messages": [response]}

    async def chat(self, message: str, context: str = "") -> str:
        """
        Send a message and get a response.

        Args:
            message: User message
            context: Optional context for the conversation

        Returns:
            Agent response
        """
        initial_state = {
            "messages": [HumanMessage(content=message)],
            "context": context,
        }
        result = await self.graph.ainvoke(initial_state)
        return result["messages"][-1].content
