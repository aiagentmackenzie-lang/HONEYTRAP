"""
HONEYTRAP AI Emulator — Dynamic response generation using Ollama.

Generates plausible service responses for honeypot emulators,
making attackers believe they're interacting with real systems.
"""

import asyncio
import hashlib
import json
import logging
import time
from collections import OrderedDict
from typing import Any

import httpx
from pydantic import BaseModel, Field

logger = logging.getLogger("honeytrap.ai")

# ─── Configuration ────────────────────────────────────────────────────────────

OLLAMA_BASE_URL = "http://localhost:11434"
OLLAMA_MODEL = "llama3.2:3b"
CACHE_MAX_SIZE = 512
CACHE_TTL_SECONDS = 3600  # 1 hour
REQUEST_TIMEOUT = 30.0

# ─── Models ────────────────────────────────────────────────────────────────────


class EmulationRequest(BaseModel):
    """Request for AI-generated service response."""

    service: str = Field(..., description="Service name: ssh, http, ftp, redis, smb")
    protocol: str = Field(default="tcp", description="Protocol: tcp or udp")
    context: dict[str, Any] = Field(
        default_factory=dict,
        description="Attack context: commands, requests, banners, etc.",
    )
    attacker_profile: dict[str, Any] = Field(
        default_factory=dict,
        description="Known attacker info: IP, tool signatures, behavior patterns.",
    )
    temperature: float = Field(default=0.7, ge=0.0, le=2.0)
    max_tokens: int = Field(default=256, ge=1, le=2048)


class EmulationResponse(BaseModel):
    """AI-generated service response."""

    response: str
    service: str
    model: str
    cached: bool
    latency_ms: float
    intent: str = Field(default="unknown", description="Detected attacker intent")
    confidence: float = Field(
        default=0.0, ge=0.0, le=1.0, description="Intent classification confidence"
    )


class IntentClassification(BaseModel):
    """Classification of attacker intent."""

    intent: str
    confidence: float
    indicators: list[str]


# ─── Response Cache ───────────────────────────────────────────────────────────


class LRUCache:
    """Least-recently-used cache for AI responses."""

    def __init__(self, max_size: int = CACHE_MAX_SIZE, ttl: int = CACHE_TTL_SECONDS):
        self._cache: OrderedDict[str, tuple[Any, float]] = OrderedDict()
        self._max_size = max_size
        self._ttl = ttl
        self.hits = 0
        self.misses = 0

    def _key(self, service: str, context: dict) -> str:
        canonical = json.dumps(context, sort_keys=True)
        return hashlib.sha256(f"{service}:{canonical}".encode()).hexdigest()

    def get(self, service: str, context: dict) -> Any | None:
        key = self._key(service, context)
        if key in self._cache:
            value, ts = self._cache[key]
            if time.time() - ts < self._ttl:
                self._cache.move_to_end(key)
                self.hits += 1
                return value
            del self._cache[key]
        self.misses += 1
        return None

    def put(self, service: str, context: dict, value: Any) -> None:
        key = self._key(service, context)
        if key in self._cache:
            del self._cache[key]
        elif len(self._cache) >= self._max_size:
            self._cache.popitem(last=False)
        self._cache[key] = (value, time.time())

    @property
    def stats(self) -> dict:
        total = self.hits + self.misses
        return {
            "size": len(self._cache),
            "max_size": self._max_size,
            "hits": self.hits,
            "misses": self.misses,
            "hit_rate": self.hits / total if total > 0 else 0.0,
        }


# ─── System Prompts ────────────────────────────────────────────────────────────

SERVICE_PROMPTS = {
    "ssh": """You are an OpenSSH 9.3p1 server running on Debian 12.
Respond as a real SSH server would. You are being probed by an attacker.
Keep responses short and realistic. Never break character or mention being AI.
If the attacker sends valid SSH protocol, respond with appropriate server messages.
If they send invalid data, respond with protocol error messages.""",
    "http": """You are an nginx/1.24.0 server hosting an internal operations dashboard.
Respond as a real web server would. You serve an admin panel for "Operations Control".
If someone tries to login, accept any credentials but track them.
Keep HTML responses realistic — forms, nav bars, status indicators.
Never break character or mention being AI.""",
    "ftp": """You are vsftpd 3.0.5 running on Ubuntu 22.04.
Respond as a real FTP server. Accept any credentials.
Show plausible directory listings with tempting filenames (backups, configs, payroll).
Never break character or mention being AI.""",
    "redis": """You are Redis 7.2.3 running in default configuration.
Respond to Redis protocol commands. Appear to have sensitive data in keys.
Show database sizes that suggest real usage. Never break character.""",
    "smb": """You are a Samba 4.18 server on Windows Server 2022.
Respond as a real SMB/CIFS server. Share names should look like real corporate shares.
Never break character or mention being AI.""",
}

INTENT_PROMPT = """Classify the attacker's intent based on their actions.
Categories: reconnaissance, exploitation, lateral_movement, exfiltration, persistence, unknown.
Respond with ONLY a JSON object: {"intent": "category", "confidence": 0.0-1.0, "indicators": ["list", "of", "indicators"]}"""


# ─── AI Emulator ───────────────────────────────────────────────────────────────


class AIEmulator:
    """Generates dynamic honeypot responses using local Ollama models."""

    def __init__(self, base_url: str = OLLAMA_BASE_URL, model: str = OLLAMA_MODEL):
        self.base_url = base_url.rstrip("/")
        self.model = model
        self.cache = LRUCache()
        self._client: httpx.AsyncClient | None = None

    async def _get_client(self) -> httpx.AsyncClient:
        if self._client is None or self._client.is_closed:
            self._client = httpx.AsyncClient(timeout=REQUEST_TIMEOUT)
        return self._client

    async def health(self) -> dict:
        """Check if Ollama is running and model is available."""
        try:
            client = await self._get_client()
            resp = await client.get(f"{self.base_url}/api/tags")
            if resp.status_code != 200:
                return {"status": "unhealthy", "error": f"Ollama returned {resp.status_code}"}
            models = [m["name"] for m in resp.json().get("models", [])]
            available = any(self.model in m for m in models)
            return {
                "status": "healthy" if available else "degraded",
                "model": self.model,
                "model_available": available,
                "available_models": models,
                "cache_stats": self.cache.stats,
            }
        except Exception as e:
            return {"status": "unhealthy", "error": str(e)}

    async def generate(self, request: EmulationRequest) -> EmulationResponse:
        """Generate a dynamic service response using Ollama."""
        # Check cache first
        cached = self.cache.get(request.service, request.context)
        if cached is not None:
            return EmulationResponse(
                response=cached["response"],
                service=request.service,
                model=self.model,
                cached=True,
                latency_ms=0.1,
                intent=cached.get("intent", "unknown"),
                confidence=cached.get("confidence", 0.0),
            )

        # Build prompt
        system_prompt = SERVICE_PROMPTS.get(request.service, SERVICE_PROMPTS["http"])
        user_prompt = self._build_user_prompt(request)

        start = time.time()
        try:
            response_text = await self._call_ollama(system_prompt, user_prompt, request)
        except Exception as e:
            logger.error(f"Ollama call failed: {e}")
            response_text = self._fallback_response(request.service)

        latency_ms = (time.time() - start) * 1000

        # Classify intent asynchronously (non-blocking)
        intent_result = await self._classify_intent(request)

        # Cache the result
        cache_value = {
            "response": response_text,
            "intent": intent_result.intent,
            "confidence": intent_result.confidence,
        }
        self.cache.put(request.service, request.context, cache_value)

        return EmulationResponse(
            response=response_text,
            service=request.service,
            model=self.model,
            cached=False,
            latency_ms=round(latency_ms, 2),
            intent=intent_result.intent,
            confidence=intent_result.confidence,
        )

    def _build_user_prompt(self, request: EmulationRequest) -> str:
        parts = []
        if request.attacker_profile:
            parts.append(f"Attacker profile: {json.dumps(request.attacker_profile)}")
        if request.context:
            parts.append(f"Context: {json.dumps(request.context)}")
        parts.append("Generate a realistic server response.")
        return "\n".join(parts)

    async def _call_ollama(
        self, system: str, user: str, request: EmulationRequest
    ) -> str:
        client = await self._get_client()
        payload = {
            "model": self.model,
            "prompt": f"{system}\n\n{user}",
            "stream": False,
            "options": {
                "temperature": request.temperature,
                "num_predict": request.max_tokens,
            },
        }
        resp = await client.post(f"{self.base_url}/api/generate", json=payload)
        resp.raise_for_status()
        data = resp.json()
        return data.get("response", "").strip()

    async def _classify_intent(self, request: EmulationRequest) -> IntentClassification:
        """Classify attacker intent using a lightweight Ollama call."""
        context_str = json.dumps(request.context)
        attacker_str = json.dumps(request.attacker_profile)

        try:
            client = await self._get_client()
            payload = {
                "model": self.model,
                "prompt": f"{INTENT_PROMPT}\n\nAttacker actions: {context_str}\nAttacker info: {attacker_str}\n\nClassify:",
                "stream": False,
                "options": {"temperature": 0.1, "num_predict": 128},
            }
            resp = await client.post(f"{self.base_url}/api/generate", json=payload)
            resp.raise_for_status()
            raw = resp.json().get("response", "").strip()
            # Parse JSON from response
            result = json.loads(raw) if raw.startswith("{") else {"intent": "unknown", "confidence": 0.0, "indicators": []}
            return IntentClassification(
                intent=result.get("intent", "unknown"),
                confidence=result.get("confidence", 0.0),
                indicators=result.get("indicators", []),
            )
        except Exception as e:
            logger.warning(f"Intent classification failed: {e}")
            return IntentClassification(intent="unknown", confidence=0.0, indicators=[])

    def _fallback_response(self, service: str) -> str:
        """Static fallback when Ollama is unavailable."""
        fallbacks = {
            "ssh": "SSH-2.0-OpenSSH_9.3p1 Debian-1\r\n",
            "http": "HTTP/1.1 200 OK\r\nContent-Type: text/html\r\n\r\n<html><body><h1>Operations Dashboard</h1><p>System nominal.</p></body></html>",
            "ftp": "220 HONEYTRAP FTP Service ready\r\n",
            "redis": "-ERR unknown command\r\n",
            "smb": "SMB2 not supported\r\n",
        }
        return fallbacks.get(service, "Service ready.\r\n")

    async def close(self) -> None:
        if self._client and not self._client.is_closed:
            await self._client.aclose()