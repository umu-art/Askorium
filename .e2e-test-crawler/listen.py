import pika, json, os
from datetime import datetime

conn = pika.BlockingConnection(pika.URLParameters(os.getenv("AMQP_URL", "amqp://guest:guest@localhost:5672/")))
ch = conn.channel()
ch.queue_declare(queue="askorium.crawler.output", durable=True)

def on_msg(ch, method, props, body):
    event = json.loads(body)
    task_id = event.get("task_id", "unknown")
    event_type = event.get("type", "unknown")
    pages = event.get("pages") or []
    stats = event.get("stats") or {}

    print(f"[{event_type}] task={task_id} pages={len(pages)} stats={stats}")

    ts = datetime.now().strftime("%Y-%m-%d_%H-%M-%S")
    out_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), "output")
    os.makedirs(out_dir, exist_ok=True)
    filename = os.path.join(out_dir, f"{ts}_{task_id[:8]}_{event_type}.json")
    with open(filename, "w", encoding="utf-8") as f:
        json.dump(event, f, indent=2, ensure_ascii=False)
    print(f"Saved: {filename}")

    ch.basic_ack(method.delivery_tag)

ch.basic_consume("askorium.crawler.output", on_msg)
print("Waiting for crawler events...")
ch.start_consuming()
