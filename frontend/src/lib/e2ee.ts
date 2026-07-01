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

export async function registerDeviceKey(token: string, deviceId: string, publicKey: string): Promise<void> {
  const res = await fetch("/api/encryption/keys", {
    method: "POST",
    headers: {
      Authorization: `Bearer ${token}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ device_id: deviceId, public_key: publicKey }),
  });
  if (!res.ok) throw new Error("key registration failed");
}
