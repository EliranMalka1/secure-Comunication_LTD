import React, { useEffect, useState } from "react";
import { Routes, Route, Navigate, Outlet, useNavigate } from "react-router-dom";
import Register from "./pages/Register.jsx";
import Login from "./pages/Login.jsx";
import Forgot from "./pages/Forgot";
import Reset from "./pages/Reset";
import Dashboard from "./pages/Dashboard";
import { apiMe } from "./lib/api";

function Home() {
  const nav = useNavigate();
  return (
    <div className="hero">
      <div className="glass">
        <h1 className="brand">Communication_LTD</h1>
        <p className="tagline">Secure communication. Seamless experience.</p>
        <div className="actions">
          <button className="btn primary" onClick={() => nav("/register")}>Create Account</button>
          <button className="btn ghost"   onClick={() => nav("/login")}>Sign In</button>
        </div>
        <div className="footer-note"><span className="dot" /> Secure Version</div>
      </div>
      <div className="background-blur blur-1" />
      <div className="background-blur blur-2" />
      <div className="background-blur blur-3" />
    </div>
  );
}

/** Guard: renders children only if the user is authenticated (cookie/JWT valid). */
function RequireAuth() {
  const [state, setState] = useState({ checking: true, ok: false });

  useEffect(() => {
    (async () => {
      try {
        await apiMe();               // 200 => session OK
        setState({ checking: false, ok: true });
      } catch {
        setState({ checking: false, ok: false }); // 401/Network => not logged in
      }
    })();
  }, []);

  if (state.checking) {
    return (
      <div className="hero">
        <div className="glass" style={{ maxWidth: 420, textAlign: "center" }}>
          Checking session…
        </div>
      </div>
    );
  }
  return state.ok ? <Outlet /> : <Navigate to="/login" replace />;
}

export default function App() {
  return (
    <Routes>
      {/* Public routes */}
      <Route path="/" element={<Home />} />
      <Route path="/register" element={<Register />} />
      <Route path="/login" element={<Login />} />
      <Route path="/forgot" element={<Forgot />} />
      <Route path="/reset" element={<Reset />} />

      {/* Protected routes (require valid session) */}
      <Route element={<RequireAuth />}>
        <Route path="/dashboard" element={<Dashboard />} />
        {/* future:
        <Route path="/change-password" element={<ChangePassword />} />
        */}
      </Route>

      {/* Fallback */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
