import { useEffect, useState } from "react";
import { useSearchParams, useNavigate } from "react-router-dom";
import { apiResetPassword } from "../lib/api";

export default function Reset() {
  const [sp] = useSearchParams();
  const nav = useNavigate();
  const token = sp.get("token") || "";
  const [password, setPassword] = useState("");
  const [msg, setMsg] = useState({ type: "", text: "" });
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!token) {
      setMsg({ type: "error", text: "Missing token." });
    }
  }, [token]);

  const submit = async (e) => {
    e.preventDefault();
    setMsg({ type: "", text: "" });
    if (!password) {
      
      setMsg({ type: "error", text: "Please enter a stronger password." });
      return;
    }
    try {
      setLoading(true);
      await apiResetPassword({ token, newPassword: password });
      setMsg({ type: "success", text: "Password reset successfully. Redirecting to sign in..." });
      setTimeout(() => nav("/login"), 1200);
    } catch (e) {
      setMsg({ type: "error", text: e?.message || "Reset failed. The link may have expired." });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="hero">
      <div className="glass" style={{ maxWidth: 520 }}>
        <h1 className="brand" style={{ fontSize: "clamp(24px,4vw,40px)" }}>
          Set a new password
        </h1>
        <p className="tagline">Choose a strong password. You’ll use it next time you sign in.</p>

        <form onSubmit={submit} style={{ textAlign: "left", marginTop: 12 }}>
          <label style={{ display: "block", marginBottom: 6 }}>New password</label>
          <input
            name="password"
            type="password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            placeholder="••••••••••"
            className="input"
            autoComplete="new-password"
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
            <button className="btn primary" type="submit" disabled={loading || !token}>
              {loading ? "Updating..." : "Update password"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
