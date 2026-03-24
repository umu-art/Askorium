import pika, json, uuid, os

URLS = [
    # "https://en.wikipedia.org/wiki/Web_scraping",
    # "https://lenta.ru/",
    "https://www.hse.ru/ma/carbon/meddoc",
    # "https://www.hse.ru/ma/carbon/meddoc",
    # "https://ba.hse.ru/faq",
    # "https://www.hse.ru/",
    # "https://en.wikipedia.org/wiki/The_Paradox_of_Choice",
]

# Документы: url + content_type_hint в metadata
DOCUMENTS = [
    # ("https://example.com/report.pdf", "pdf"),
    # ("https://example.com/data.xlsx", "xlsx"),
    ("https://mozilla.github.io/pdf.js/web/compressed.tracemonkey-pldi-09.pdf", "pdf"),
]

conn = pika.BlockingConnection(pika.URLParameters(os.getenv("AMQP_URL", "amqp://guest:guest@localhost:5672/")))
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
        properties=pika.BasicProperties(delivery_mode=2),
    )
    print("Sent:", url)

for url, hint in DOCUMENTS:
    msg = {
        "task_id": str(uuid.uuid4()),
        "url": url,
        "metadata": {"content_type_hint": hint},
    }
    ch.basic_publish(
        exchange="",
        routing_key="askorium.render.input",
        body=json.dumps(msg),
        properties=pika.BasicProperties(delivery_mode=2),
    )
    print(f"Sent [{hint}]:", url)

conn.close()
