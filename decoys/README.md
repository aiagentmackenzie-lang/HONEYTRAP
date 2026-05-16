# Decoy Files

This directory contains planted fake credentials for honeypot detection.
These files are **not tracked in git** to prevent accidental exposure.

## Generating Decoys

Use the CLI (when implemented) or the Go token generator to create decoy files:

```bash
# Generate decoy files (future: ./honeytrap tokens generate)
# Currently, create them manually using internal/tokens package
```

## File Types

- `fake-aws-credentials.json` — Planted AWS access keys
- `fake-database-config.yml` — Fake DB connection strings
- `fake-api-key.env` — Planted environment variables

⚠️ **Warning:** Decoy files contain realistic-looking fake credentials.
Do not commit to public repos or expose to systems that might auto-scan them.