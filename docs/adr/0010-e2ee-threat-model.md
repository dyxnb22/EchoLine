# ADR 0010: End-to-End Encryption Threat Model

## Status

Accepted (design; E2EE implementation deferred to extension phase)

## Context

EchoLine currently stores messages in plaintext in Postgres. The server can read any message. This is acceptable for a demo platform but is architecturally incompatible with a privacy-first messaging product.

Users increasingly expect E2EE as a baseline (Signal, WhatsApp, iMessage all offer it). Regulators in several jurisdictions are beginning to require it for certain message categories. More critically, a single Postgres breach would expose all historical messages for all users.

Before designing a key management system (ADR 0011), we must define the threat model: **what are we protecting against, and what tradeoffs are acceptable?**

## Threat Model

### Assets

| Asset | Sensitivity | Location (current) |
|-------|-------------|-------------------|
| Message content | High | Postgres `messages.body` |
| Media files | High | MinIO object store |
| User identity (username, phone) | Medium | Postgres `users` |
| Delivery metadata (who sent to whom, when) | Medium | Postgres `deliveries` |
| Device keys (future E2EE) | Critical | Client device (never server) |

### Adversaries

| Adversary | Capability | Target |
|-----------|-----------|--------|
| External attacker | Network interception, DB breach | Message content, credentials |
| Malicious server operator | Full DB access, server-side code execution | Message content, metadata |
| Compromised device | Access to device key material | Past and future messages |
| Nation-state / legal compulsion | Warrant to server operator | Message content (metadata always at risk) |
| Passive network observer | TLS-terminated traffic analysis | Metadata (timing, size) |

### Trust Boundary

E2EE **protects message content from the server**. It does **not** protect:
- Metadata (who talks to whom, when, conversation membership)
- Delivery status (sent/delivered/read)
- Account registration information
- Traffic analysis (connection patterns)

This is consistent with Signal's threat model and is an accepted limitation of server-mediated E2EE.

### Attacks In Scope

1. **DB exfiltration**: An attacker gains read access to Postgres. Without E2EE, all message history is exposed. With E2EE, ciphertext is useless without device keys.
2. **Server-side key recovery**: The server must not be able to derive any message key. Key derivation happens only on client devices.
3. **MITM on key exchange**: The server could substitute a rogue public key during device registration. Mitigated by key transparency / Safety Numbers (out of scope for MVP E2EE).
4. **Replay attack**: An attacker replays an old encrypted message. Mitigated by the Double Ratchet's message key derivation; each message uses a unique ephemeral key.

### Attacks Out of Scope (MVP)

- Forward secrecy for group conversations (complex ratchet state management across thousands of members)
- Sealed sender / traffic analysis resistance
- Key transparency / audit log (Safety Numbers equivalent)
- Backup encryption key recovery (iCloud/Google Drive backup)

## Decision

Design E2EE using the **Signal Protocol (Double Ratchet + X3DH)**:

- Clients generate and store identity keys, signed pre-keys, and one-time pre-keys on device.
- The server acts as a **key distribution server**: it stores public keys only, never private keys.
- Message encryption/decryption happens entirely on client devices.
- The server stores **ciphertext blobs** in the `messages.encrypted_body` column (E2EE) alongside `messages.body = NULL` (to prevent accidental plaintext leakage).
- Server cannot read message content. It can read metadata (sender, recipient, timestamp, size).

## Implementation Files

- `docs/adr/0011-e2ee-key-management.md` — key storage, rotation, pre-key bundle API
- `backend/migrations/` — add `encrypted_body` column, make `body` nullable
- `backend/internal/api/keys.go` _(planned)_ — pre-key bundle upload/fetch API
- `docs/security-checklist.md` — E2EE deployment checklist

## Consequences

**Positive:**
- DB breach does not expose message content.
- Server operator (including employees) cannot read messages.
- Legally, server cannot comply with warrants for message content (only metadata).

**Negative:**
- Key backup/recovery is the user's responsibility; lost device = lost message history.
- Group E2EE with large groups (>100 members) has high key distribution overhead.
- Search over encrypted messages requires client-side index or plaintext excerpt stored server-side (both have tradeoffs).
- E2EE breaks server-side content moderation (spam, CSAM scanning).

## Interview Talking Points

- **What does E2EE actually protect?** "It protects message content from the server and from anyone who breaches the server. It does not protect metadata — the server still knows who is talking to whom. Signal calls this 'sealed sender' and adds extra mechanisms for it."
- **Why Signal Protocol?** "X3DH gives us asynchronous key exchange (sender can send to an offline recipient using their pre-key bundle). Double Ratchet gives forward secrecy and break-in recovery."
- **Search problem**: "This is the hardest E2EE UX problem. Options: client-side search index (complex, must be rebuilt on new device), server-side encrypted search (SGX/trusted execution), or no search. Telegram's approach is no E2EE for cloud chats, which enables search but breaks confidentiality."
- **Content moderation**: "Apple's CSAM scanning controversy illustrates the tension. E2EE and server-side moderation are fundamentally incompatible. This is a product and legal decision, not just a technical one."
