from dotenv import load_dotenv
import os

load_dotenv()

# --- RAG API ---
RAG_BASE_URL: str = os.environ["RAG_BASE_URL"]
RAG_SOURCE_ID: str = os.environ["RAG_SOURCE_ID"]
RAG_MODE: str = os.getenv("RAG_MODE", "deep")
RAG_POLL_INTERVAL: float = float(os.getenv("RAG_POLL_INTERVAL", "1.0"))
RAG_TIMEOUT: float = float(os.getenv("RAG_TIMEOUT", "60.0"))
RAG_CONCURRENCY: int = int(os.getenv("RAG_CONCURRENCY", "5"))

# --- Dataset ---
DATASET_PATH: str = os.getenv("DATASET_PATH", "dataset.jsonl")

# --- LLM Judge (OpenRouter) ---
OPENROUTER_API_KEY: str = os.environ["OPENROUTER_API_KEY"]
OPENROUTER_BASE_URL: str = "https://openrouter.ai/api/v1"
JUDGE_MODEL: str = os.getenv("JUDGE_MODEL", "anthropic/claude-haiku-4-6")

# --- Embeddings (OpenRouter) ---
EMBEDDING_MODEL: str = os.getenv("EMBEDDING_MODEL", "google/gemini-embedding-exp-03-07")

# --- Retrieval метрики ---
RETRIEVAL_K: int = int(os.getenv("RETRIEVAL_K", "5"))
RETRIEVAL_THRESHOLD: float = float(os.getenv("RETRIEVAL_THRESHOLD", "0.5"))
