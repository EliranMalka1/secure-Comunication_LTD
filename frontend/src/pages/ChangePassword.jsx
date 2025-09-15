import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiPasswordChange } from "../lib/api";

export default function ChangePassword() {
  const nav = useNavigate();
  const [form, setForm] = useState({ old: "", next: "", confirm: "" });
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState({ type: "", text: "" });

  const onChange = (e) => setForm({ ...form, [e.target.name]: e.target.value });

  const validate = () => {
    if (!form.old || !form.next || !form.confirm) return "Please fill all fields.";
    if (form.next !== form.confirm) return "New password and confirmation do not match.";
    if (form.next.length < 10) return "New password looks too short."; 
    return "";
  };

  const onSubmit = async (e) => {
    e.preventDefault();
    setMsg({ type: "", text: "" });
    const err = validate();
    if (err) return setMsg({ type: "error", text: err });

    try {
      setLoading(true);
      await apiPasswordChange({ oldPassword: form.old, newPassword: form.next });
      setMsg({ type: "success", text: "Check your email to confirm the change." });
      setTimeout(() => nav("/dashboard"), 1200);
    } catch (e) {
      setMsg({ type: "error", text: e?.message || "Request failed" });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="hero">
      <div className="glass" style={{ maxWidth: 520 }}>
        <h1 className="brand" style={{ fontSize: "clamp(24px,4vw,40px)" }}>Change Password</h1>
        <p className="tagline">Enter your current password and a new password.</p>

        <form onSubmit={onSubmit} style={{ textAlign: "left", marginTop: 12 }}>
          <label style={{ display: "block", marginBottom: 6 }}>Current password</label>
          <input name="old" type="password" className="input" value={form.old} onChange={onChange} />

          <label style={{ display: "block", margin: "14px 0 6px" }}>New password</label>
          <input name="next" type="password" className="input" value={form.next} onChange={onChange} />

          <label style={{ display: "block", margin: "14px 0 6px" }}>Confirm new password</label>
          <input name="confirm" type="password" className="input" value={form.confirm} onChange={onChange} />

          {msg.text && (
            <div style={{
              marginTop: 14, padding: "10px 12px", borderRadius: 10,
              background: msg.type === "error" ? "rgba(255,0,0,0.12)" : "rgba(0,255,120,0.12)",
              border: "1px solid rgba(255,255,255,0.18)",
            }}>
              {msg.text}
            </div>
          )}

          <div className="actions" style={{ marginTop: 18, justifyContent: "flex-end" }}>
            <button className="btn ghost" type="button" onClick={() => nav("/dashboard")}>Cancel</button>
            <button className="btn primary" type="submit" disabled={loading}>
              {loading ? "Sending…" : "Request Change"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
