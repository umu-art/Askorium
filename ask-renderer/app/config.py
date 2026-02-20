import os


class Config:
    AMQP_URL: str = os.getenv("AMQP_URL", "amqp://guest:guest@localhost:5672/")
    INPUT_QUEUE: str = os.getenv("INPUT_QUEUE", "render.input")
    OUTPUT_QUEUE: str = os.getenv("OUTPUT_QUEUE", "render.output")
    MAX_CONCURRENT_PAGES: int = int(os.getenv("MAX_CONCURRENT_PAGES", "5"))
    DEFAULT_TIMEOUT_MS: int = int(os.getenv("DEFAULT_TIMEOUT_MS", "15000"))
    MAX_RETRIES: int = int(os.getenv("MAX_RETRIES", "2"))
    MESSAGE_TTL_MS: int = int(os.getenv("MESSAGE_TTL_MS", "60000"))
    RECONNECT_INTERVAL_S: int = int(os.getenv("RECONNECT_INTERVAL_S", "5"))
    HEALTH_PORT: int = int(os.getenv("HEALTH_PORT", "8080"))
    LOG_LEVEL: str = os.getenv("LOG_LEVEL", "INFO").upper()


config = Config()
