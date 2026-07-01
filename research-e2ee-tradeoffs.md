# Research: E2EE Tradeoffs — Signal Protocol vs MLS vs Pairwise RSA

## Purpose

This document analyses the tradeoffs between the three main approaches to end-to-end encryption in messaging systems, with specific focus on the constraints and design decisions relevant to EchoLine.

---

## 1. Threat Model

Before choosing an E2EE scheme, we define what we are protecting against:

| Threat | Protected by E2EE? |
|--------|-------------------|
| Passive eavesdropping (TLS interception) | Yes |
| Compromised EchoLine server (server operator reads DB) | Yes |
| Subpoena / legal access to server | Yes (only ciphertext available) |
| Compromised user device | No (device holds plaintext) |
| Man-in-the-middle at key exchange (server substitutes keys) | Partially (key verification required) |
| Traffic analysis (who talks to whom) | No — metadata is visible to server |
| Future decryption of stored ciphertext ("harvest now, decrypt later") | Depends on forward secrecy |

EchoLine's E2EE goal: protect **message content** at rest and in transit, even if the server database is fully compromised. **Metadata** (sender, recipient, timestamp, message count) remains visible to the server.

---

## 2. Option A: Signal Protocol (X3DH + Double Ratchet)

### How It Works

**X3DH (Extended Triple Diffie-Hellman)**: establishes a shared secret between two parties using a key bundle published to the server (identity key, signed pre-key, one-time pre-keys). No online interaction required — the sender can initiate even if the recipient is offline.

**Double Ratchet**: maintains a chain of symmetric keys. Every message advances the ratchet; every DH reply generates a new ratchet root. This provides:

- **Forward secrecy**: compromise of today's keys cannot decrypt past messages.
- **Break-in recovery**: after a key compromise, a new DH ratchet step generates fresh keys. Within ~2 message exchanges, confidentiality is restored.

### Multi-Device

Signal Protocol sessions are **per-device** (not per-user). If Alice has two phones, Bob must establish two independent X3DH sessions — one with each device. The sender encrypts the message twice (once per recipient device). For N recipient devices, the sender sends N ciphertexts.

This scales linearly with device count. WhatsApp (which uses Signal Protocol) caps devices per account at 4 linked devices.

### Group E2EE

Signal's Sender Keys protocol is used for groups:

1. Each group member generates a "Sender Key" — a chain key + signature key pair.
2. Each member distributes their Sender Key to all other group members (N² key exchange for a group of N).
3. When Alice sends a group message, she encrypts once with her Sender Key. Each recipient uses Alice's Sender Key to decrypt.
4. On member add/remove, all Sender Keys are rotated.

**Limitation**: N² key exchange for large groups. WhatsApp solves this with a dedicated "key distribution" message type. For groups > 1000, Sender Key distribution is expensive.

### Pros/Cons

| | Signal Protocol |
|-|----------------|
| ✅ Forward secrecy | Per-message via Double Ratchet |
| ✅ Break-in recovery | Yes, after new DH ratchet |
| ✅ Battle-tested | Signal, WhatsApp, Facebook Messenger |
| ✅ Mature libraries | libsignal (C, Java, Swift, TypeScript) |
| ✗ Multi-device | Per-device sessions (linear message copies) |
| ✗ Large groups | N² key exchange for Sender Key distribution |
| ✗ Key transparency | Relies on trust in key server (no built-in transparency log) |

---

## 3. Option B: MLS (Messaging Layer Security, RFC 9420)

### How It Works

MLS is an IETF standard (RFC 9420, 2023) designed specifically for **group E2EE at scale**. It uses a **binary tree (TreeKEM)** for key agreement:

- Group state is represented as a binary tree of Diffie-Hellman key pairs.
- Adding/removing a member requires updating only the O(log N) nodes on the path from that leaf to the root.
- This reduces key exchange from O(N²) to **O(log N)** per group mutation.

Every group operation (join, leave, update) generates a new **Epoch** with fresh group keys. Forward secrecy is maintained between epochs.

### Multi-Device

MLS treats each device as a separate leaf node in the group tree. Multi-device is native: adding a new device is just adding a leaf. Removing a device (revocation) is a group operation that updates O(log N) nodes.

### Pros/Cons

| | MLS |
|-|-----|
| ✅ O(log N) group operations | vs O(N²) for Signal Sender Keys |
| ✅ Native multi-device | Each device is a leaf node |
| ✅ Standard (RFC 9420) | Interoperability potential |
| ✅ Post-compromise security | Fresh epoch keys after member rotation |
| ✗ Complexity | Much more complex to implement correctly |
| ✗ Library maturity | OpenMLS (Rust) is good; Go/TypeScript implementations are younger |
| ✗ Ordering requirements | MLS requires strict epoch ordering; out-of-order messages require careful buffering |
| ✗ Server-assisted | Requires a "Delivery Service" that stores the current group state (tree) |

---

## 4. Option C: Pairwise RSA / ECIES (Naive Approach)

### How It Works

Each user generates an RSA-2048 or EC key pair. To send a message, the sender:
1. Fetches the recipient's public key from the server.
2. Encrypts a random AES session key with the recipient's RSA public key.
3. Encrypts the message with the AES session key.
4. Sends both encrypted blobs.

### Pros/Cons

| | Pairwise RSA/ECIES |
|-|-------------------|
| ✅ Simple to implement | Well-understood primitives |
| ✗ No forward secrecy | Compromise of private key decrypts all past messages |
| ✗ No break-in recovery | Once compromised, always compromised |
| ✗ Multi-device: complex | Must encrypt for each device separately; no standard protocol |
| ✗ Group: impractical | N copies of the message (one per member) |

**Verdict**: Acceptable only for very low-security use cases. Not appropriate for a production messaging system.

---

## 5. Comparison Table

| Property | Signal Protocol | MLS (RFC 9420) | Pairwise RSA |
|----------|----------------|----------------|--------------|
| Forward secrecy | Per-message | Per-epoch | None |
| Break-in recovery | Yes (DH ratchet) | Yes (epoch rotation) | No |
| Multi-device overhead | O(devices) per send | O(log N) per op | O(devices) per send |
| Group scaling | O(N²) key exchange | O(log N) per op | O(N) per send |
| Implementation maturity | High (libsignal) | Medium (OpenMLS) | High (but insecure) |
| IETF standard | Informal spec | RFC 9420 | No |
| Key transparency | None built-in | None built-in (IETF TLS planned) | None |

---

## 6. EchoLine Decision

**Phase 1**: Signal Protocol (X3DH + Double Ratchet) for 1:1 conversations. Sender Keys for small groups (< 100 members). Rationale: mature libraries (libsignal-protocol-typescript for browser, libsignal for mobile), well-understood operational characteristics.

**Phase 2**: Monitor MLS ecosystem maturity. If OpenMLS + a Go/TypeScript MLS client library reach production readiness by the time EchoLine needs large-group E2EE (> 1000 members), migrate to MLS for new groups. Signal Sender Keys remain for existing groups.

**Key Transparency**: In both phases, users are advised to verify **safety numbers** (key fingerprints) out-of-band to protect against a compromised key server substituting keys. A key transparency log (similar to the CT log for TLS certificates) is on the research roadmap.

---

## 7. Interview Talking Points

> **"Why Signal Protocol over MLS?"**
> "MLS is theoretically superior for large groups — O(log N) vs O(N²) for key operations. But libsignal has 10+ years of production hardening across billions of users. Our groups are < 500 members for MVP, so O(N²) is fine. We revisit MLS when groups grow beyond that threshold."

> **"What does 'forward secrecy' mean in practice?"**
> "If an attacker records all encrypted traffic today and steals a user's device keys next year, can they decrypt those old messages? With a static RSA key: yes. With Signal's Double Ratchet: no — each message uses a derived key that is deleted after use. Even if you get today's keys, you can't decrypt yesterday's messages."

> **"What metadata does E2EE hide?"**
> "E2EE protects message *content*. The server still knows who talked to whom, when, and how many messages. For a subpoena, the government can still get the conversation graph — just not the content. Full metadata protection requires something like Tor + sealed sender (Signal's approach), which is beyond MVP scope."
