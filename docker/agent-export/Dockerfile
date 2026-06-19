FROM python:3.12-alpine

WORKDIR /app

COPY server.py .

ENV AGENT_EXPORT_DIR=/exports
ENV AGENT_EXPORT_HOST=0.0.0.0
ENV AGENT_EXPORT_PORT=2030

EXPOSE 2030

CMD ["python", "server.py"]
