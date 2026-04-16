"""
HONEYTRAP AI Emulator — FastAPI server exposing AI emulation endpoints.

Provides:
  POST /ai-response   — Generate dynamic honeypot response
  GET  /ai/health     — Check Ollama connectivity
  GET  /ai/cache      — Cache statistics
"""

import logging
import sys

import uvicorn
from fastapi import FastAPI, HTTPException

from emulator import AIEmulator, EmulationRequest, EmulationResponse

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(name)s] %(levelname)s: %(message)s",
)

app = FastAPI(
    title="HONEYTRAP AI Emulator",
    description="AI-powered dynamic response generation for honeypot services",
    version="0.2.0",
)

emulator = AIEmulator()


@app.on_event("shutdown")
async def shutdown():
    await emulator.close()


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