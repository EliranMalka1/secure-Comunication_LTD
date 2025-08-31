import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiLogin, apiLoginMFA } from "../lib/api";

export default function Login() {
  const nav = useNavigate();

  // שני מצבים: "password" (ברירת מחדל) ואז "otp"
  const [step, setStep] = useState("password");
  const [form, setForm] = useState({ id: "", password: "", code: "" });
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState({ type: "", text: "" });
  const [otpMeta, setOtpMeta] = useState({ expiresIn: 0, method: "" });

  const onChange = (e) => setForm({ ...form, [e.target.name]: e.target.value });

  const looksLikeEmail = (s) => s.includes("@") && s.includes(".");
  const looksLikeUsername = (s) => /^[a-zA-Z0-9._-]{3,}$/.test(s);

  const validatePasswordStep = () => {
    if (!form.id.trim() || !form.password.trim()) return "Please fill in all required fields.";
    const id = form.id.trim();
    if (!(looksLikeEmail(id) || looksLikeUsername(id))) return "Enter a valid email or username.";
    return "";
  };

  const validateOTP = () => {
    if (!form.code.trim()) return "Please enter the 6-digit code from your email.";
    if (!/^\d{6}$/.test(form.code.trim())) return "The code must be exactly 6 digits.";
    return "";
  };

  // שלב 1: סיסמה
  const onSubmitPassword = async (e) => {
    e.preventDefault();
    setMsg({ type: "", text: "" });

    const err = validatePasswordStep();
    if (err) return setMsg({ type: "error", text: err });

    try {
      setLoading(true);
      const data = await apiLogin({
        id: form.id.trim(),
        password: form.password,
      });

      // אם השרת דורש MFA – נעבור לשלב הקוד
      if (data?.mfa_required) {
        setOtpMeta({ expiresIn: data.expires_in ?? 10, method: data.method ?? "email_otp" });
        setForm((f) => ({ ...f, code: "" }));
        setStep("otp");
        setMsg({
          type: "success",
          text: `We sent a verification code to your email. It expires in ${data.expires_in ?? 10} minutes.`,
        });
        return;
      }

      // אחרת (אם בעתיד תאפשר לוגין בלי 2FA)
      setMsg({ type: "success", text: "Logged in successfully." });
      setTimeout(() => nav("/"), 500);
    } catch (e) {
      setMsg({ type: "error", text: e?.message || "Login failed. Please try again." });
    } finally {
      setLoading(false);
    }
  };

  // שלב 2: שליחת OTP
  const onSubmitOTP = async (e) => {
    e.preventDefault();
    setMsg({ type: "", text: "" });

    const err = validateOTP();
    if (err) return setMsg({ type: "error", text: err });

    try {
      setLoading(true);
      await apiLoginMFA({
        id: form.id.trim(),
        code: form.code.trim(),
      });

      setMsg({ type: "success", text: "Verification successful. Redirecting..." });
      setTimeout(() => nav("/"), 500);
    } catch (e) {
      setMsg({ type: "error", text: e?.message || "Invalid code. Please try again." });
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="hero">
      <div className="glass" style={{ maxWidth: 520 }}>
        <h1 className="brand" style={{ fontSize: "clamp(24px,4vw,40px)" }}>
          {step === "password" ? "Sign In" : "Verify Code"}
        </h1>
        <p className="tagline">
          {step === "password"
            ? "Welcome back. Please sign in to continue."
            : `Enter the 6-digit code we sent to your email. ${otpMeta.expiresIn ? `Expires in ${otpMeta.expiresIn} min.` : ""}`}
        </p>

        {step === "password" ? (
          <form onSubmit={onSubmitPassword} style={{ textAlign: "left", marginTop: 12 }}>
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
                {loading ? "Signing in..." : "Continue"}
              </button>
            </div>
          </form>
        ) : (
          <form onSubmit={onSubmitOTP} style={{ textAlign: "left", marginTop: 12 }}>
            <label style={{ display: "block", marginBottom: 6 }}>Verification code</label>
            <input
              name="code"
              type="text"
              inputMode="numeric"
              pattern="\d*"
              maxLength={6}
              value={form.code}
              onChange={onChange}
              placeholder="123456"
              className="input"
              autoComplete="one-time-code"
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

            <div className="actions" style={{ marginTop: 18, justifyContent: "space-between" }}>
              <button className="btn ghost" type="button" onClick={() => setStep("password")}>
                Back
              </button>
              <div style={{ display: "flex", gap: 10 }}>
                {}
                <button className="btn primary" type="submit" disabled={loading}>
                  {loading ? "Verifying..." : "Verify"}
                </button>
              </div>
            </div>
          </form>
        )}

        {step === "password" && (
          <div className="footer-note">
            <span className="dot" /> Don’t have an account?
            <button className="btn ghost" onClick={() => nav("/register")}>Create one</button>
          </div>
        )}
      </div>
    </div>
  );
}
