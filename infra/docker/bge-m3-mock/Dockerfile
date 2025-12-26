FROM python:3.11-slim

WORKDIR /app

ENV PYTHONUNBUFFERED=1
ENV PORT=8091

RUN pip install --no-cache-dir flask gunicorn

COPY server.py .

EXPOSE 8091

CMD ["gunicorn", "-b", "0.0.0.0:8091", "server:app"]
