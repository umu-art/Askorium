import asyncio
import json
import logging
from asyncio import StreamReader, StreamWriter

from collections.abc import Callable

from prometheus_client import generate_latest, CONTENT_TYPE_LATEST

logger = logging.getLogger("health")


class HealthServer:
    def __init__(self):
        self._checks: dict[str, Callable] = {}
        self._server: asyncio.Server | None = None

    def register(self, name: str, check: Callable) -> None:
        self._checks[name] = check

    async def _handle(self, reader: StreamReader, writer: StreamWriter) -> None:
        request_line = await reader.readline()
        while True:
            line = await reader.readline()
            if line in (b"\r\n", b"\n", b""):
                break

        path = request_line.decode().split(" ")[1] if b" " in request_line else "/"

        if path == "/metrics":
            body = generate_latest()
            response = (
                f"HTTP/1.1 200 OK\r\n"
                f"Content-Type: {CONTENT_TYPE_LATEST}\r\n"
                f"Content-Length: {len(body)}\r\n"
                f"\r\n"
            ).encode() + body
        else:
            results = {}
            healthy = True
            for name, check in self._checks.items():
                try:
                    ok = await check()
                    results[name] = "ok" if ok else "unhealthy"
                    if not ok:
                        healthy = False
                except Exception as e:
                    results[name] = f"error: {e}"
                    healthy = False

            status = 200 if healthy else 503
            body = json.dumps({"status": "ok" if healthy else "unhealthy", "checks": results})
            response = (
                f"HTTP/1.1 {status} {'OK' if healthy else 'Service Unavailable'}\r\n"
                f"Content-Type: application/json\r\n"
                f"Content-Length: {len(body)}\r\n"
                f"\r\n"
                f"{body}"
            ).encode()

        writer.write(response)
        await writer.drain()
        writer.close()

    async def start(self, port: int) -> None:
        self._server = await asyncio.start_server(self._handle, "0.0.0.0", port)
        logger.info("health server listening on :%d", port)

    async def stop(self) -> None:
        if self._server:
            self._server.close()
            await self._server.wait_closed()
