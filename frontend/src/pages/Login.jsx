import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiLogin, apiLoginMFA, apiForgotPassword } from "../lib/api";

export default function Login() {
  const nav = useNavigate();

  // Two states: "password" (default) and then "otp"
  const [step, setStep] = useState("password");
  const [form, setForm] = useState({ id: "", password: "", code: "" });
  const [loading, setLoading] = useState(false);
  const [msg, setMsg] = useState({ type: "", text: "" });
  const [otpMeta, setOtpMeta] = useState({ expiresIn: 0, method: "" });

  // Forgot password UI state
  const [showForgot, setShowForgot] = useState(false);
  const [forgotEmail, setForgotEmail] = useState("");
  const [forgotLoading, setForgotLoading] = useState(false);
  const [forgotMsg, setForgotMsg] = useState({ type: "", text: "" });

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

  // Step 1: Password
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

      // If the server requires MFA - move to the code step
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

      // Otherwise (if in the future you allow login without 2FA)
      setMsg({ type: "success", text: "Logged in successfully." });
      setTimeout(() => nav("/dashboard"), 500);
    } catch (e) {
      setMsg({ type: "error", text: e?.message || "Login failed. Please try again." });
    } finally {
      setLoading(false);
    }
  };

  // Step 2:  Send OTP
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

  // Forgot Password submit
  const onSubmitForgot = async (e) => {
    e.preventDefault();
    setForgotMsg({ type: "", text: "" });

    const email = forgotEmail.trim();
    if (!email || !looksLikeEmail(email)) {
      setForgotMsg({ type: "error", text: "Please enter a valid email address." });
      return;
    }

    try {
      setForgotLoading(true);
      await apiForgotPassword(email);
      setForgotMsg({
        type: "success",
        text: "If this email exists, a reset link has been sent. Please check your inbox.",
      });
    } catch (err) {
      setForgotMsg({ type: "error", text: err?.message || "Request failed. Try again later." });
    } finally {
      setForgotLoading(false);
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
            : `Enter the 6-digit code we sent to your email. ${
                otpMeta.expiresIn ? `Expires in ${otpMeta.expiresIn} min.` : ""
              }`}
        </p>

        {step === "password" ? (
          <>
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

              {/* Forgot password toggle */}
              <div style={{ marginTop: 10 }}>
                <button
                  type="button"
                  className="btn ghost"
                  onClick={() => {
                    setShowForgot((v) => !v);
                    setForgotMsg({ type: "", text: "" });
                  }}
                >
                  {showForgot ? "Hide forgot password" : "Forgot your password?"}
                </button>
              </div>

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

            {/* Forgot password panel */}
            {showForgot && (
              <form onSubmit={onSubmitForgot} style={{ textAlign: "left", marginTop: 18 }}>
                <hr style={{ opacity: 0.15, margin: "12px 0 16px" }} />
                <h3 style={{ margin: "0 0 8px" }}>Reset your password</h3>
                <p className="tagline" style={{ marginTop: 0 }}>
                  Enter your account email and we’ll send a reset link.
                </p>
                <label style={{ display: "block", marginBottom: 6 }}>Email</label>
                <input
                  name="forgotEmail"
                  type="email"
                  value={forgotEmail}
                  onChange={(e) => setForgotEmail(e.target.value)}
                  placeholder="you@example.com"
                  className="input"
                  autoComplete="email"
                />

                {forgotMsg.text && (
                  <div
                    style={{
                      marginTop: 14,
                      padding: "10px 12px",
                      borderRadius: 10,
                      background:
                        forgotMsg.type === "error" ? "rgba(255,0,0,0.12)" : "rgba(0,255,120,0.12)",
                      border: "1px solid rgba(255,255,255,0.18)",
                    }}
                  >
                    {forgotMsg.text}
                  </div>
                )}

                <div className="actions" style={{ marginTop: 14, justifyContent: "flex-end" }}>
                  <button className="btn primary" type="submit" disabled={forgotLoading}>
                    {forgotLoading ? "Sending..." : "Send reset link"}
                  </button>
                </div>
              </form>
            )}
          </>
        ) : (
          <form onSubmit={onSubmitOTP} style={{ textAlign: "left", marginTop: 12 }}>
            <label style={{ display: "block", marginBottom: 6 }}>Verification code</label>
            <input
              name="code"
              type="text"
              inputMode="numeric"
              pattern="[0-9]*"
              minLength={6}
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
            <button className="btn ghost" onClick={() => nav("/register")}>
              Create one
            </button>
          </div>
        )}
      </div>
    </div>
  );
}
