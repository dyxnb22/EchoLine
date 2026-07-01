# API Gateway Prototype

EchoLine modular monolith routes through a single `api` process today. This directory holds a **reverse-proxy skeleton** for future service split (T441+).

## Purpose

- Terminate TLS at the edge.
- Route `/api/*` to the monolith API.
- Route `/ws` to the realtime gateway.
- Reserve paths for future microservices (`/svc/auth`, `/svc/message`, …).

## Local usage

```bash
docker compose --profile gateway up nginx-gateway
```

Or point any nginx/Caddy instance at `nginx.conf`.

## Tradeoffs

| Approach | Pros | Cons |
|----------|------|------|
| Monolith direct | Simple, one deploy unit | Harder per-service scaling |
| Gateway + monolith | Path-based routing, TLS centralization | Extra hop, config drift |
| Full microservices | Independent scale/deploy | Distributed tracing, S2S auth required |

## Related

- ADR 0027 — API gateway prototype
- ADR 0028 — Microservices boundary
- `docs/extensions-roadmap.md` — microservices section
