import os


class Config:
    AMQP_URL: str = os.getenv("AMQP_URL", "amqp://guest:guest@localhost:5672/")
    INPUT_QUEUE: str = os.getenv("INPUT_QUEUE", "askorium.render.output")
    OUTPUT_QUEUE: str = os.getenv("OUTPUT_QUEUE", "askorium.parser.output")
    PREFETCH_COUNT: int = int(os.getenv("PREFETCH_COUNT", "10"))
    MESSAGE_TTL_MS: int = int(os.getenv("MESSAGE_TTL_MS", "60000"))
    RECONNECT_INTERVAL_S: int = int(os.getenv("RECONNECT_INTERVAL_S", "5"))
    LOG_LEVEL: str = os.getenv("LOG_LEVEL", "INFO").upper()


config = Config()
