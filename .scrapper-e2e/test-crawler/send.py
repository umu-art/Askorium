import pika, json, uuid, os

# Домен и стартовые URL для обхода
TASKS = [
    {
        "domain": "hse.ru",
        "seed_urls": ["https://hse.ru/"],
        "max_pages": 20,
        "max_depth": 5,
    },
]

conn = pika.BlockingConnection(pika.URLParameters(os.getenv("AMQP_URL", "amqp://guest:guest@localhost:5672/")))
ch = conn.channel()
ch.queue_declare(queue="askorium.crawler.input", durable=True, passive=True)

for t in TASKS:
    msg = {
        "task_id": str(uuid.uuid4()),
        "domain": t["domain"],
        "seed_urls": t["seed_urls"],
        "options": {
            "max_pages": t["max_pages"],
            "max_depth": t["max_depth"],
        },
    }
    ch.basic_publish(
        exchange="",
        routing_key="askorium.crawler.input",
        body=json.dumps(msg),
        properties=pika.BasicProperties(delivery_mode=2),
    )
    print(f"Sent: {t['domain']} (max_pages={t['max_pages']}), task_id={msg['task_id']}")

conn.close()
