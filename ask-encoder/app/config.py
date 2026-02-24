from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    embedding_model: str = "BAAI/bge-m3"
    reranker_model: str = "BAAI/bge-reranker-v2-m3"
    host: str = "0.0.0.0"
    port: int = 8000
    max_concurrent_requests: int = 2

    model_config = {"env_prefix": "ENCODER_"}


settings = Settings()
