import { useState } from "react";
import { apiForgotPassword } from "../lib/api";

export default function Forgot() {
  const [email, setEmail] = useState("");
  const [msg, setMsg] = useState({ type: "", text: "" });
  const [loading, setLoading] = useState(false);

  const submit = async (e) => {
    e.preventDefault();
    setMsg({ type: "", text: "" });
    if (!email.trim() || !email.includes("@")) {
      setMsg({ type: "error", text: "Please enter a valid email." });
      return;
    }
    try {
      setLoading(true);
      await apiForgotPassword(email.trim());
      setMsg({
        type: "success",
        text: "If this email exists, a reset link has been sent.",
      });
    } catch (e) {
  // Even on error, return a generic message
      setMsg({
        type: "success",
        text: "If this email exists, a reset link has been sent.",
      });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="hero">
      <div className="glass" style={{ maxWidth: 520 }}>
        <h1 className="brand" style={{ fontSize: "clamp(24px,4vw,40px)" }}>
          Forgot Password
        </h1>
        <p className="tagline">Enter your email to receive a reset link.</p>

        <form onSubmit={submit} style={{ textAlign: "left", marginTop: 12 }}>
          <label style={{ display: "block", marginBottom: 6 }}>Email</label>
          <input
            name="email"
            type="email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            placeholder="you@example.com"
            className="input"
            autoComplete="email"
          />

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
            <button className="btn primary" type="submit" disabled={loading}>
              {loading ? "Sending..." : "Send reset link"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
