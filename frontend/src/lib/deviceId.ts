const STORAGE_KEY = "echoline_device";

/** Stable per-browser device id used for sync, WS, push, and E2EE registration. */
export function getOrCreateDeviceId(): string {
  let id = localStorage.getItem(STORAGE_KEY);
  if (!id) {
    id = crypto.randomUUID();
    localStorage.setItem(STORAGE_KEY, id);
  }
  return id;
}
