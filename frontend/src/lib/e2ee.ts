/**
 * E2EE client prototype — XOR obfuscation for demo only.
 * Production would use Signal Protocol / WebCrypto AES-GCM.
 */
export function encryptPayload(plaintext: string, deviceKey: string): string {
  const key = deviceKey.padEnd(plaintext.length, "0").slice(0, plaintext.length);
  return btoa(
    Array.from(plaintext)
      .map((c, i) => String.fromCharCode(c.charCodeAt(0) ^ key.charCodeAt(i % key.length)))
      .join(""),
  );
}

export function decryptPayload(ciphertext: string, deviceKey: string): string {
  const raw = atob(ciphertext);
  const key = deviceKey.padEnd(raw.length, "0").slice(0, raw.length);
  return Array.from(raw)
    .map((c, i) => String.fromCharCode(c.charCodeAt(0) ^ key.charCodeAt(i % key.length)))
    .join("");
}

import { authedRequest, parseResponse } from "../api/http";

export async function registerDeviceKey(token: string, deviceId: string, publicKey: string): Promise<void> {
  const res = await authedRequest(token, "/api/encryption/keys", {
    method: "POST",
    body: JSON.stringify({ device_id: deviceId, public_key: publicKey }),
  });
  await parseResponse(res, "key registration failed");
}
