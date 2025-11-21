import hashlib
import os
import random
from typing import List, Union

from flask import Flask, jsonify, request

app = Flask(__name__)

EMBEDDING_DIM = int(os.environ.get("EMBEDDING_DIM", "1024"))


def generate_embedding(text: str) -> List[float]:
    """Create a deterministic pseudo-embedding so tests behave consistently."""
    seed = int(hashlib.sha256(text.encode("utf-8")).hexdigest(), 16)
    rng = random.Random(seed)
    return [rng.uniform(-1.0, 1.0) for _ in range(EMBEDDING_DIM)]


def normalize_inputs(payload: dict) -> List[str]:
    inputs: Union[str, List[str]] = payload.get("inputs", [])
    if isinstance(inputs, str):
        return [inputs]
    if not isinstance(inputs, list):
        return []
    return [str(item) for item in inputs]


@app.get("/health")
def health() -> tuple[str, int]:
    return "ok", 200


@app.get("/info")
def info():
    return jsonify({"model_id": "BAAI/bge-m3"})


@app.post("/embed")
def embed():
    payload = request.get_json(force=True, silent=True) or {}
    inputs = normalize_inputs(payload)
    embeddings = [generate_embedding(text) for text in inputs]
    return jsonify(embeddings)


@app.post("/embed_sparse")
def embed_sparse():
    # The client only checks that the endpoint responds with valid JSON.
    payload = request.get_json(force=True, silent=True) or {}
    inputs = normalize_inputs(payload)
    empty_sparse = [[{"index": 0, "value": 0.0}] for _ in inputs]
    return jsonify(empty_sparse)


if __name__ == "__main__":
    app.run(host="0.0.0.0", port=int(os.environ.get("PORT", "8091")))
