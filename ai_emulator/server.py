"""
HONEYTRAP AI Emulator — FastAPI server exposing AI emulation endpoints.

Provides:
  POST /ai-response   — Generate dynamic honeypot response
  GET  /ai/health     — Check Ollama connectivity
  GET  /ai/cache      — Cache statistics
"""

import logging
import os
import sys
from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import JSONResponse

from emulator import AIEmulator, EmulationRequest, EmulationResponse

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(name)s] %(levelname)s: %(message)s",
)

API_KEY = os.environ.get("HONEYTRAP_API_KEY", "")


@asynccontextmanager
async def lifespan(app: FastAPI):
    # Startup — nothing to do
    yield
    # Shutdown
    await emulator.close()


app = FastAPI(
    title="HONEYTRAP AI Emulator",
    description="AI-powered dynamic response generation for honeypot services",
    version="0.2.0",
    lifespan=lifespan,
)

emulator = AIEmulator()


@app.middleware("http")
async def auth_middleware(request: Request, call_next):
    # Skip auth for health/docs endpoints
    if request.url.path in ("/ai/health", "/docs", "/openapi.json", "/redoc"):
        return await call_next(request)

    # Skip auth if API_KEY is not set (dev mode)
    if not API_KEY:
        return await call_next(request)

    auth = request.headers.get("Authorization", "")
    if auth != f"Bearer {API_KEY}":
        return JSONResponse(status_code=401, content={"error": "Unauthorized"})
    return await call_next(request)


@app.post("/ai-response", response_model=EmulationResponse)
async def generate_response(request: EmulationRequest) -> EmulationResponse:
    """Generate a dynamic service response using AI emulation."""
    try:
        return await emulator.generate(request)
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))


@app.get("/ai/health")
async def health_check():
    """Check Ollama connectivity and model availability."""
    return await emulator.health()


@app.get("/ai/cache")
async def cache_stats():
    """Get response cache statistics."""
    return emulator.cache.stats


if __name__ == "__main__":
    port = int(sys.argv[1]) if len(sys.argv) > 1 else 8443
    uvicorn.run(app, host="0.0.0.0", port=port)