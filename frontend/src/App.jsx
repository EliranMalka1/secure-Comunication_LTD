import React, { useEffect, useState } from "react";
import { Routes, Route, Navigate, Outlet, useNavigate } from "react-router-dom";
import Register from "./pages/Register.jsx";
import Login from "./pages/Login.jsx";
import Forgot from "./pages/Forgot";
import Reset from "./pages/Reset";
import Dashboard from "./pages/Dashboard";
import ChangePassword from "./pages/ChangePassword.jsx";
import { apiMe } from "./lib/api";
import CustomerNew from "./pages/CustomerNew";


/** Home: If there is a session - automatically navigates to the dashboard; otherwise it displays the home page. */
function Home() {
  const nav = useNavigate();
  const [checking, setChecking] = useState(true);

  useEffect(() => {
    (async () => {
      try {
        await apiMe(); // There is a valid cookie/session
        nav("/dashboard", { replace: true });
      } catch {
        // Not connected - stay on the home page
      } finally {
        setChecking(false);
      }
    })();
  }, [nav]);

  if (checking) {
    return (
      <div className="hero">
        <div className="glass" style={{ maxWidth: 420, textAlign: "center" }}>
          Checking session…
        </div>
      </div>
    );
  }

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

/** Guard: Allows access to routes only if there is a valid session (cookie/JWT). */
function RequireAuth() {
  const [state, setState] = useState({ checking: true, ok: false });

  useEffect(() => {
    (async () => {
      try {
        await apiMe(); // 200 => connected
        setState({ checking: false, ok: true });
      } catch {
        setState({ checking: false, ok: false });
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

/** Guard: Blocks Login/Register/Forgot/Reset pages if already logged in. */
function PublicOnly() {
  const [state, setState] = useState({ checking: true, loggedIn: false });

  useEffect(() => {
    (async () => {
      try {
        await apiMe(); // If successful - already connected
        setState({ checking: false, loggedIn: true });
      } catch {
        setState({ checking: false, loggedIn: false });
      }
    })();
  }, []);

  if (state.checking) return null; // Minimal spinner - not mandatory to display

  return state.loggedIn ? <Navigate to="/dashboard" replace /> : <Outlet />;
}

export default function App() {
  return (
    <Routes>
      {/* Public root (with redirect if connected) */}
      <Route path="/" element={<Home />} />

      {/* Public pages - blocked if already connected */}
      <Route element={<PublicOnly />}>
        <Route path="/register" element={<Register />} />
        <Route path="/login" element={<Login />} />
        <Route path="/forgot" element={<Forgot />} />
        <Route path="/reset" element={<Reset />} />
      </Route>

      {/* Protected pages - require a session */}
      <Route element={<RequireAuth />}>
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/change-password" element={<ChangePassword />} />
        <Route path="/dashboard" element={<Dashboard />} />
        <Route path="/customers/new" element={<CustomerNew />} />
      </Route>

      {/* Fallback */}
      <Route path="*" element={<Navigate to="/" replace />} />
    </Routes>
  );
}
