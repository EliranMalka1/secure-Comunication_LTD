const BASE_URL = import.meta.env.VITE_API_URL || "http://localhost:8080";

export async function apiRegister(payload) {
  const res = await fetch(`${BASE_URL}/api/register`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    
    body: JSON.stringify(payload),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data?.error || data?.message || "Registration failed");
  }
  return data;
}
