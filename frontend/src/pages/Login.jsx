import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiLogin } from "../lib/api";

export default function Login() {
  const nav = useNavigate();
  const [form, setForm] = useState({ id: "", password: "", otp: "" });
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState({ type: "", text: "" });

  const onChange = (e) => setForm({ ...form, [e.target.name]: e.target.value });

  const looksLikeEmail = (s) => s.includes("@") && s.includes(".");
  const looksLikeUsername = (s) => /^[a-zA-Z0-9._-]{3,}$/.test(s);

  const validate = () => {
    if (!form.id.trim() || !form.password.trim()) return "Please fill in all required fields.";
    const id = form.id.trim();
    if (!(looksLikeEmail(id) || looksLikeUsername(id))) return "Enter a valid email or username.";
    return "";
  };

  const onSubmit = async (e) => {
    e.preventDefault();
    setMsg({ type: "", text: "" });

    const err = validate();
    if (err) return setMsg({ type: "error", text: err });

    try {
      setLoading(true);
      await apiLogin({
        id: form.id.trim(),
        password: form.password,
        otp: form.otp || undefined,
      });
      setMsg({ type: "success", text: "Logged in successfully." });
      setTimeout(() => nav("/"), 600);
    } catch (e) {
      setMsg({ type: "error", text: e?.message || "Network error. Please try again." });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="hero">
      <div className="glass" style={{ maxWidth: 520 }}>
        <h1 className="brand" style={{ fontSize: "clamp(24px,4vw,40px)" }}>
          Sign In
        </h1>
        <p className="tagline">Welcome back. Please sign in to continue.</p>

        <form onSubmit={onSubmit} style={{ textAlign: "left", marginTop: 12 }}>
          <label style={{ display: "block", marginBottom: 6 }}>Email or Username</label>
          <input
            name="id"
            type="text"
            value={form.id}
            onChange={onChange}
            placeholder="you@example.com or yourname"
            className="input"
            autoComplete="username"
          />

          <label style={{ display: "block", margin: "14px 0 6px" }}>Password</label>
          <input
            name="password"
            type="password"
            value={form.password}
            onChange={onChange}
            placeholder="••••••••"
            className="input"
            autoComplete="current-password"
          />

          {/* OTP field optional for future (currently can be hidden or shown as optional) */}
          {/* <label style={{ display: "block", margin: "14px 0 6px" }}>2FA Code (optional)</label>
          <input
            name="otp"
            type="text"
            inputMode="numeric"
            pattern="[0-9]*"
            value={form.otp}
            onChange={onChange}
            placeholder="123456"
            className="input"
            autoComplete="one-time-code"
          /> */}

          {msg.text && (
            <div
              style={{
                marginTop: 14,
                padding: "10px 12px",
                borderRadius: 10,
                background: msg.type === "error" ? "rgba(255,0,0,0.12)" : "rgba(0,255,120,0.12)",
                border: "1px solid rgba(255,255,255,0.18)",
              }}
            >
              {msg.text}
            </div>
          )}

          <div className="actions" style={{ marginTop: 18, justifyContent: "flex-end" }}>
            <button className="btn ghost" type="button" onClick={() => nav("/")}>
              Cancel
            </button>
            <button className="btn primary" type="submit" disabled={loading}>
              {loading ? "Signing in..." : "Sign In"}
            </button>
          </div>
        </form>

        <div className="footer-note">
          <span className="dot" /> Don’t have an account?
          <button className="btn ghost" onClick={() => nav("/register")}>Create one</button>
        </div>
      </div>
    </div>
  );
}
