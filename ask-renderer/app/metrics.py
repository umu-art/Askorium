from prometheus_client import Counter, Gauge, Histogram

renderer_messages_total = Counter(
    "renderer_messages_total",
    "Total messages processed",
    ["status"],
)

renderer_errors_total = Counter(
    "renderer_errors_total",
    "Errors by error code",
    ["code"],
)

renderer_retries_total = Counter(
    "renderer_retries_total",
    "Total retry attempts",
)

renderer_duration_seconds = Histogram(
    "renderer_duration_seconds",
    "Render duration in seconds",
)

renderer_active_tasks = Gauge(
    "renderer_active_tasks",
    "Currently processing tasks",
)
