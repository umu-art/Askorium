import pika, json, uuid

URLS = [
    # "https://en.wikipedia.org/wiki/Web_scraping",
    # "https://lenta.ru/",
    "https://www.hse.ru/ma/carbon/meddoc",
]

conn = pika.BlockingConnection(pika.URLParameters("amqp://guest:guest@localhost:5672/"))
ch = conn.channel()
ch.queue_declare(queue="askorium.render.input", durable=True, passive=True)

for url in URLS:
    msg = {
        "task_id": str(uuid.uuid4()),
        "url": url,
        "timeout_ms": 15000,
    }
    ch.basic_publish(
        exchange="",
        routing_key="askorium.render.input",
        body=json.dumps(msg),
        properties=pika.BasicProperties(delivery_mode=2),  # persistent
    )
    print("Sent:", url)

conn.close()