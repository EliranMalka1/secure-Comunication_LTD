import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiCreateCustomer } from "../lib/api";

export default function CustomerNew() {
  const nav = useNavigate();
  const [form, setForm] = useState({ name: "", email: "", phone: "", notes: "" });
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState({ type: "", text: "" });
  const [createdName, setCreatedName] = useState("");

  const onChange = (e) => setForm({ ...form, [e.target.name]: e.target.value });

  const validate = () => {
    if (!form.name.trim() || !form.email.trim()) return "Name & email are required.";
    if (!(form.email.includes("@") && form.email.includes("."))) return "Invalid email.";
    return "";
  };

  const onSubmit = async (e) => {
    e.preventDefault();
    setMsg({ type: "", text: "" });
    const err = validate();
    if (err) return setMsg({ type: "error", text: err });

    try {
      setLoading(true);
      const data = await apiCreateCustomer({
        name: form.name.trim(),
        email: form.email.trim(),
        phone: form.phone.trim(),
        notes: form.notes,
      });
      setCreatedName(data.name);
      setMsg({ type: "success", text: `Customer created: ${data.name}` });
      
       setTimeout(() => nav("/dashboard"), 700);
    } catch (e) {
      setMsg({ type: "error", text: e?.message || "Create failed" });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="hero">
      <div className="glass" style={{ maxWidth: 640, textAlign: "left" }}>
        <h1 className="brand" style={{ fontSize: "clamp(24px,4vw,40px)" }}>New Customer</h1>
        <p className="tagline">Add a customer and we’ll show their name upon success.</p>

        <form onSubmit={onSubmit} style={{ marginTop: 16 }}>
          <label>Name</label>
          <input name="name" className="input" value={form.name} onChange={onChange} placeholder="Acme Ltd" />

          <label style={{ marginTop: 12 }}>Email</label>
          <input name="email" className="input" value={form.email} onChange={onChange} placeholder="contact@acme.com" />

          <label style={{ marginTop: 12 }}>Phone (optional)</label>
          <input name="phone" className="input" value={form.phone} onChange={onChange} placeholder="+972-50-0000000" />

          <label style={{ marginTop: 12 }}>Notes (optional)</label>
          <textarea name="notes" className="input" rows={3} value={form.notes} onChange={onChange} placeholder="Any notes..." />

          {msg.text && (
            <div style={{
              marginTop: 14, padding: "10px 12px", borderRadius: 10,
              background: msg.type === "error" ? "rgba(255,0,0,0.12)" : "rgba(0,255,120,0.12)",
              border: "1px solid rgba(255,255,255,0.18)"
            }}>
              {msg.text}
            </div>
          )}

          {createdName && (
            <div style={{ marginTop: 10, opacity: 0.9 }}>
              <strong>Created:</strong> {createdName}
            </div>
          )}

          <div className="actions" style={{ marginTop: 18, justifyContent: "flex-end" }}>
            <button className="btn ghost" type="button" onClick={() => nav("/dashboard")}>Cancel</button>
            <button className="btn primary" type="submit" disabled={loading}>
              {loading ? "Saving..." : "Create"}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
