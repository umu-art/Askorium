import time

import pika, json, os
from datetime import datetime

conn = pika.BlockingConnection(pika.URLParameters("amqp://guest:guest@localhost:5672/"))
ch = conn.channel()
ch.queue_declare(queue="askorium.parser.output", durable=True, passive=True)

def url_to_filename(url: str) -> str:
    import re
    url = re.sub(r"^https?://", "", url)
    url = re.sub(r"[^\w\-.]", "_", url)
    url = re.sub(r"_+", "_", url).strip("_")
    return url[:120]

def on_msg(ch, method, props, body):
    resp = json.loads(body)
    page_url = (resp.get("page") or {}).get("url") or resp.get("task_id", "unknown")
    ts = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")
    out_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "output")
    os.makedirs(out_dir, exist_ok=True)
    filename = os.path.join(out_dir, f"{ts}_{url_to_filename(page_url)}.json")
    with open(filename, "w", encoding="utf-8") as f:
        json.dump(resp, f, indent=2, ensure_ascii=False)
    print(f"Saved: {filename}")
    ch.basic_ack(method.delivery_tag)

ch.basic_consume("askorium.parser.output", on_msg)
print("Waiting for messages...")
ch.start_consuming()