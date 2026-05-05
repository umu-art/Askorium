from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    embedding_model: str = "BAAI/bge-m3"
    reranker_model: str = "BAAI/bge-reranker-v2-m3"
    max_concurrent_requests: int = 2
    temporal_url: str = "localhost:7233"
    task_queue: str = "askorium-encoder"

    model_config = {"env_prefix": "ENCODER_"}


settings = Settings()
