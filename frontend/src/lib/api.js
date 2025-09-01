const BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

// Helpers
async function parseJson(res) {
  try { return await res.json(); } catch { return {}; }
}
async function assertOk(res) {
  if (!res.ok) {
    const data = await parseJson(res);
    const message = data?.error || data?.message || `Request failed (${res.status})`;
    throw new Error(message);
  }
  return res;
}

async function post(path, body, { withCredentials = true } = {}) {
  const res = await fetch(`${BASE_URL}${path}`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: withCredentials ? "include" : "omit",
    body: JSON.stringify(body),
  });
  await assertOk(res);
  return parseJson(res);
}

async function get(path) {
  const res = await fetch(`${BASE_URL}${path}`, {
    method: "GET",
    credentials: "include",
  });
  await assertOk(res);
  return parseJson(res);
}

// === Public API ===

// Registration should not send cookies
export async function apiRegister(payload) {
  const res = await fetch(`${BASE_URL}/api/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "omit",
    body: JSON.stringify(payload),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data?.error || data?.message || "Registration failed");
  }
  return data;
}

// Step 1: password login (server will return { mfa_required: true } if 2FA is on)
export async function apiLogin(payload) {
  return post("/api/login", payload);
}

// Step 2: submit OTP code
export async function apiLoginMFA({ id, code }) {
  return post("/api/login/mfa", { id, code });
}

// Optional: logout + whoami
export async function apiLogout() {
  return post("/api/logout", {}, { withCredentials: true });
}
export async function apiMe() {
  return get("/api/me");
}

export async function apiForgotPassword(email) {
  // No need for cookies
  const res = await fetch(`${BASE_URL}/api/password/forgot`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "omit",
    body: JSON.stringify({ email }),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data?.error || "Request failed");
  return data;
}

export async function apiResetPassword({ token, newPassword }) {
  const res = await fetch(`${BASE_URL}/api/password/reset`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "omit",
    body: JSON.stringify({ token, new_password: newPassword }),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data?.error || "Request failed");
  return data;
}

export async function apiPasswordChange({ oldPassword, newPassword }) {
  // requires valid session cookie
  const res = await fetch(`${BASE_URL}/api/password/change`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ old_password: oldPassword, new_password: newPassword })
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data?.error || "Request failed");
  return data;
}

export async function apiCreateCustomer(payload) {
  const res = await fetch(`${BASE_URL}/api/customers`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include", // requires valid session cookie
    body: JSON.stringify(payload),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data?.error || "Create customer failed");
  return data; // { id, name }
}

export async function apiSearchCustomers({ q = "", page = 1, size = 10 } = {}) {
  const params = new URLSearchParams();
  if (q) params.set("q", q);
  params.set("page", String(page));
  params.set("size", String(size));

  const res = await fetch(`${BASE_URL}/api/customers/search?${params.toString()}`, {
    method: "GET",
    credentials: "include",
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data?.error || "Search failed");
  return data; // { items, page, size, total }
}
