# ADR 0026: E2EE Key Bundle Registry

## Status

Accepted (prototype)

## Context

End-to-end encryption requires per-device public key registration before ciphertext exchange.

## Decision

- Table `encryption_key_bundles` (migration 00012).
- REST: `POST/GET /api/encryption/keys` — register/list device public keys.
- Server stores public keys only; no private key material on server.

## Out of Scope (prototype)

Double ratchet, pre-key bundles, sealed sender, encrypted search.

## Files

- `backend/internal/encryption/handler.go`
- `docs/encryption-prototype.md`
