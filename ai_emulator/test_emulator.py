"""HONEYTRAP AI Emulator — Test suite."""

import asyncio
import json
from unittest.mock import AsyncMock, patch

import pytest
from emulator import AIEmulator, EmulationRequest, LRUCache


# ─── LRU Cache Tests ──────────────────────────────────────────────────────────


class TestLRUCache:
    def test_put_and_get(self):
        cache = LRUCache(max_size=10)
        cache.put("ssh", {"banner": "test"}, "response_text")
        result = cache.get("ssh", {"banner": "test"})
        assert result == "response_text"

    def test_cache_miss(self):
        cache = LRUCache()
        result = cache.get("http", {"path": "/nonexistent"})
        assert result is None

    def test_eviction(self):
        cache = LRUCache(max_size=3)
        cache.put("ssh", {"i": 1}, "a")
        cache.put("http", {"i": 2}, "b")
        cache.put("ftp", {"i": 3}, "c")
        cache.put("redis", {"i": 4}, "d")  # evicts ssh
        assert cache.get("ssh", {"i": 1}) is None
        assert cache.get("redis", {"i": 4}) == "d"

    def test_stats(self):
        cache = LRUCache()
        cache.put("ssh", {"cmd": "ls"}, "response")
        cache.get("ssh", {"cmd": "ls"})  # hit
        cache.get("http", {"path": "/"})  # miss
        stats = cache.stats
        assert stats["hits"] == 1
        assert stats["misses"] == 1
        assert stats["hit_rate"] == 0.5


# ─── Emulation Request Tests ──────────────────────────────────────────────────


class TestEmulationRequest:
    def test_defaults(self):
        req = EmulationRequest(service="ssh", context={"banner": "SSH-2.0-OpenSSH"})
        assert req.protocol == "tcp"
        assert req.temperature == 0.7
        assert req.max_tokens == 256

    def test_custom_params(self):
        req = EmulationRequest(
            service="http",
            context={"method": "POST", "path": "/login"},
            attacker_profile={"ip": "10.0.0.1"},
            temperature=0.3,
            max_tokens=512,
        )
        assert req.service == "http"
        assert req.attacker_profile["ip"] == "10.0.0.1"


# ─── AI Emulator Tests ────────────────────────────────────────────────────────


class TestAIEmulator:
    def test_fallback_response(self):
        emulator = AIEmulator()
        assert "OpenSSH" in emulator._fallback_response("ssh")
        assert "200 OK" in emulator._fallback_response("http")
        assert "FTP" in emulator._fallback_response("ftp")
        assert "ERR" in emulator._fallback_response("redis")
        assert "ready" in emulator._fallback_response("unknown")

    @pytest.mark.asyncio
    async def test_generate_with_ollama_down(self):
        """When Ollama is unavailable, fallback should be used."""
        emulator = AIEmulator(base_url="http://localhost:99999")
        request = EmulationRequest(
            service="ssh",
            context={"banner": "SSH-2.0-libssh2"},
        )
        # This should fall back gracefully since Ollama isn't running
        response = await emulator.generate(request)
        assert response.service == "ssh"
        # Response will be fallback since Ollama is down
        assert len(response.response) > 0