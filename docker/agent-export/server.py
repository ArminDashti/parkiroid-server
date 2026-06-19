from __future__ import annotations

import datetime
import json
import os
import uuid
from http.server import BaseHTTPRequestHandler, HTTPServer

EXPORT_DIR = os.environ.get("AGENT_EXPORT_DIR", "/exports")
HOST = os.environ.get("AGENT_EXPORT_HOST", "0.0.0.0")
PORT = int(os.environ.get("AGENT_EXPORT_PORT", "2030"))


class AgentExportHandler(BaseHTTPRequestHandler):
    def do_GET(self) -> None:
        if self.path == "/health":
            self._send_json(200, {"status": "ok"})
            return

        self.send_error(404)

    def do_POST(self) -> None:
        if self.path != "/api/agent-responses":
            self.send_error(404)
            return

        length = int(self.headers.get("Content-Length", "0"))
        raw = self.rfile.read(length)

        try:
            payload = json.loads(raw.decode("utf-8"))
        except (UnicodeDecodeError, json.JSONDecodeError):
            self.send_error(400, "invalid json")
            return

        text = str(payload.get("text", "")).strip()
        if not text:
            self.send_error(400, "text is required")
            return

        conversation_id = str(payload.get("conversation_id", "unknown"))
        exported_at = str(
            payload.get("exported_at", datetime.datetime.now(datetime.UTC).isoformat())
        )
        timestamp = datetime.datetime.now(datetime.UTC).strftime("%Y%m%d-%H%M%S")
        filename = f"{timestamp}_{conversation_id[:8]}_{uuid.uuid4().hex[:6]}.md"
        filepath = os.path.join(EXPORT_DIR, filename)

        lines = [
            "# Agent Response",
            "",
            "| Field | Value |",
            "|-------|-------|",
            f"| Exported At | {exported_at} |",
            f"| Conversation | {conversation_id} |",
        ]

        generation_id = payload.get("generation_id")
        if generation_id:
            lines.append(f"| Generation | {generation_id} |")

        model = payload.get("model")
        if model:
            lines.append(f"| Model | {model} |")

        transcript_path = payload.get("transcript_path")
        if transcript_path:
            lines.append(f"| Transcript | {transcript_path} |")

        lines.extend(["", "---", "", text, ""])

        os.makedirs(EXPORT_DIR, exist_ok=True)
        with open(filepath, "w", encoding="utf-8") as handle:
            handle.write("\n".join(lines))

        self._send_json(
            201,
            {
                "file": filename,
                "path": filepath,
            },
        )

    def _send_json(self, status: int, payload: dict) -> None:
        body = json.dumps(payload).encode("utf-8")
        self.send_response(status)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)

    def log_message(self, format: str, *args) -> None:
        return


def main() -> None:
    os.makedirs(EXPORT_DIR, exist_ok=True)
    server = HTTPServer((HOST, PORT), AgentExportHandler)
    print(f"agent-export listening on {HOST}:{PORT}, writing to {EXPORT_DIR}")
    server.serve_forever()


if __name__ == "__main__":
    main()
