const BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

// Helper: returns JSON if available, returns {} otherwise
async function parseJson(res) {
  try { return await res.json(); } catch { return {}; }
}

// Helper: throws error with a consistent message
async function assertOk(res) {
  if (!res.ok) {
    const data = await parseJson(res);
    const message = data?.error || data?.message || `Request failed (${res.status})`;
    throw new Error(message);
  }
  return res;
}

async function post(path, body) {
  const res = await fetch(`${BASE_URL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
  credentials: "include",          // Important for httpOnly cookies
    body: JSON.stringify(body),
  });
  await assertOk(res);
  return parseJson(res);
}

async function get(path) {
  const res = await fetch(`${BASE_URL}${path}`, {
    method: "GET",
  credentials: "include",          // Important for httpOnly cookies
  });
  await assertOk(res);
  return parseJson(res);
}

// === Public API ===

export async function apiRegister(payload) {
  const res = await fetch(`${BASE_URL}/api/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
  credentials: "omit",       // Do not send cookies during registration
    body: JSON.stringify(payload),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data?.error || data?.message || "Registration failed");
  }
  return data;
}


export async function apiLogin(payload) {
  // payload: { id, password, otp? }
  return post("/api/login", payload);
}

export async function apiLogout() {
  // Server will return 200 and clear cookie
  return post("/api/logout", {});
}

export async function apiMe() {
  // Server will return user details based on the cookie (when implemented)
  return get("/api/me");
}
