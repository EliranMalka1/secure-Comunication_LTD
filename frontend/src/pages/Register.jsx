import React, { useState } from "react";
import { apiRegister } from "../lib/api";

export default function Register() {
  const [form, setForm] = useState({ username: "", email: "", password: "", confirm: "" });
  const [submitting, setSubmitting] = useState(false);
  const [msg, setMsg] = useState({ type: "", text: "" });

  const onChange = (e) => {
    const { name, value } = e.target;
    setForm((f) => ({ ...f, [name]: value }));
    setMsg({ type: "", text: "" });
  };

  const isEmail = (s) => /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(s);
  const basicOk =
    form.username.trim() &&
    isEmail(form.email) &&
    form.password &&
    form.confirm &&
    form.password === form.confirm;

  const onSubmit = async (e) => {
    e.preventDefault();
    if (!basicOk) return;
    setSubmitting(true);
    setMsg({ type: "", text: "" });
    try {
      await apiRegister({
        username: form.username.trim(),
        email: form.email.trim(),
        password: form.password,
      });
      setMsg({ type: "ok", text: "Account created. And verify mail sent to your email." });
      setForm({ username: "", email: "", password: "", confirm: "" });
    } catch (err) {
      setMsg({ type: "error", text: err.message || "Registration failed" });
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <div className="hero">
      <div className="glass" style={{ maxWidth: 520, width: "92vw", textAlign: "left" }}>
        <h2 className="brand" style={{ fontSize: "clamp(22px,4vw,34px)", marginBottom: 6 }}>
          Create your account
        </h2>
        <p className="tagline" style={{ marginBottom: 22 }}>
          Basic client-side checks. Password policy is enforced on the server.
        </p>

        <form onSubmit={onSubmit} noValidate>
          <Field label="Username">
            <input name="username" value={form.username} onChange={onChange}
              placeholder="e.g., eli123" required className="input" autoComplete="username" />
          </Field>

          <Field label="Email">
            <input type="email" name="email" value={form.email} onChange={onChange}
              placeholder="you@example.com" required className="input" autoComplete="email" />
            {form.email && !isEmail(form.email) && (
              <Note type="warn">Please enter a valid email.</Note>
            )}
          </Field>

          <Field label="Password">
            <input type="password" name="password" value={form.password} onChange={onChange}
              placeholder="Password" required className="input" autoComplete="new-password" />
          </Field>

          <Field label="Confirm Password">
            <input type="password" name="confirm" value={form.confirm} onChange={onChange}
              placeholder="Repeat password" required className="input" autoComplete="new-password" />
            {form.confirm && form.confirm !== form.password && (
              <Note type="warn">Passwords do not match.</Note>
            )}
          </Field>

          {msg.text && <Note type={msg.type}>{msg.text}</Note>}

          <div className="actions" style={{ justifyContent: "flex-end" }}>
            <button className="btn ghost" type="button" onClick={() => window.history.back()}>
              Back
            </button>
            <button className="btn primary" disabled={!basicOk || submitting}>
              {submitting ? "Creating..." : "Create Account"}
            </button>
          </div>
        </form>
      </div>

      <div className="background-blur blur-1" />
      <div className="background-blur blur-2" />
      <div className="background-blur blur-3" />

      <style>{`
        .input {
          width: 100%;
          border-radius: 12px;
          border: 1px solid rgba(255,255,255,0.15);
          background: rgba(255,255,255,0.04);
          outline: none; color: #e8e8f0; padding: 12px 14px;
        }
        .input:focus { border-color: #6c8bff; box-shadow: 0 0 0 2px rgba(108,139,255,.25); }
      `}</style>
    </div>
  );
}

function Field({ label, children }) {
  return (
    <div style={{ marginBottom: 14 }}>
      <label style={{ display: "block", fontSize: 13, opacity: 0.9, marginBottom: 6 }}>{label}</label>
      {children}
    </div>
  );
}

function Note({ type = "ok", children }) {
  const colors = {
    ok: "rgba(60, 227, 122, .2)",
    warn: "rgba(247, 201, 72, .15)",
    error: "rgba(255, 107, 107, .18)",
  };
  return (
    <div style={{ background: colors[type], padding: "8px 10px", borderRadius: 10, marginTop: 10, fontSize: 13 }}>
      {children}
    </div>
  );
}
