# End-to-End Encryption Prototype — EchoLine

## Problem

Users expect that their private messages cannot be read by EchoLine's servers or administrators. Standard server-side encryption (data at rest + TLS in transit) protects against external attackers but not against a compromised server or a subpoena. End-to-end encryption (E2EE) ensures that only the sender and recipient can decrypt message content.

See also: ADR 0010 (E2EE Threat Model), ADR 0011 (Key Management).

## Tradeoff

| Protocol | Perfect Forward Secrecy | Multi-device | Group E2EE | Complexity |
|----------|------------------------|--------------|------------|------------|
| Signal Protocol (Double Ratchet + X3DH) | Yes | Per-device session | Sender Keys | High |
| MLS (RFC 9420) | Yes | Native | Native, efficient | Very High |
| Pairwise RSA (naive) | No | Manual | Impractical | Low |

**Decision**: Implement Signal Protocol (X3DH + Double Ratchet) for 1:1 E2EE. Group E2EE uses the Signal Sender Keys protocol. MLS is tracked as a future upgrade path (see ADR 0010). See ADR 0011 for key storage strategy.

## Key Concepts

### X3DH Initial Key Exchange

Before the first message, the sender fetches the recipient's public keys from the EchoLine key server and computes a shared secret:

```
Sender computes:
  DH1 = DH(IK_sender,    SPK_recipient)
  DH2 = DH(EK_sender,    IK_recipient)
  DH3 = DH(EK_sender,    SPK_recipient)
  DH4 = DH(EK_sender,    OPK_recipient)   // if one-time key available
  master_secret = KDF(DH1 || DH2 || DH3 || DH4)
```

Where:
- `IK` = Identity Key (long-term)
- `SPK` = Signed Pre-Key (rotated monthly)
- `EK` = Ephemeral Key (one-time per exchange)
- `OPK` = One-Time Pre-Key (consumed once)

### Double Ratchet (Ongoing Messages)

After X3DH, each message advances the ratchet:

```
Sending key chain:  KDF(chain_key) → message_key, next_chain_key
Receiving key chain: same, mirrored
DH ratchet: every reply triggers a new DH exchange → fresh chain keys
```

This provides **forward secrecy** (compromise of today's keys doesn't decrypt past messages) and **break-in recovery** (new DH ratchet after compromise rotates keys).

## Implementation Files

- `frontend/src/lib/e2ee/x3dh.ts` — X3DH key exchange (using WebCrypto API)
- `frontend/src/lib/e2ee/ratchet.ts` — Double Ratchet state machine
- `frontend/src/lib/e2ee/keystore.ts` — IndexedDB key storage, never leaves device
- `backend/internal/keyserver/handler.go` — publish/fetch pre-keys (opaque to server)
- `backend/migrations/00017_key_bundles.sql` — pre-key storage schema
- `docs/adr/0010-e2ee-threat-model.md` — threat model
- `docs/adr/0011-e2ee-key-management.md` — key lifecycle

## Schema (Key Server — Server Stores No Plaintext Keys)

```sql
CREATE TABLE key_bundles (
    user_id       UUID NOT NULL REFERENCES users(id),
    device_id     UUID NOT NULL REFERENCES devices(id),
    ik_public     BYTEA NOT NULL,    -- identity key (public only)
    spk_public    BYTEA NOT NULL,    -- signed pre-key (public only)
    spk_signature BYTEA NOT NULL,    -- IK signature over SPK
    opk_public    BYTEA[],           -- one-time pre-keys (public only)
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_id, device_id)
);
```

The server stores **only public keys**. Private keys never leave the device.

## Message Format (E2EE)

```json
{
  "client_msg_id": "...",
  "conversation_id": "...",
  "e2ee": {
    "version": 1,
    "ephemeral_key": "<base64>",
    "ciphertext": "<base64>",
    "mac": "<base64>"
  }
}
```

The `text` field is absent. The server stores and forwards the opaque `e2ee` object. The recipient's client decrypts it using the ratchet state.

## Key Fingerprint Verification

To prevent MITM via key server compromise, users can compare key fingerprints out-of-band:

```
Your safety number with Alice:
  38291 04821 93847 20381
  48291 09182 73648 20918
```

This is the same mechanism used by Signal ("Safety Numbers") and WhatsApp ("Security Code").

## Testing

```bash
# Frontend E2EE unit tests (Vitest)
cd frontend
npm run test -- --run src/lib/e2ee/

# Integration: Alice sends encrypted message, Bob decrypts
RUN_E2EE_TEST=1 npm run test -- --run src/tests/e2ee-round-trip.test.ts
```

## Interview Angle

> "E2EE is implemented entirely on the client. The server is a dumb relay for opaque ciphertext — it cannot read any message content. The key engineering challenge is multi-device: if Alice has two phones, both need to decrypt the message. We solve this with Signal's Sender Keys protocol for groups and per-device X3DH sessions for 1:1 conversations."
