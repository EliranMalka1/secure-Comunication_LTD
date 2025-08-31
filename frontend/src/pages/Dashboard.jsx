import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiMe, apiLogout } from "../lib/api";

function Clock() {
  const [now, setNow] = useState(new Date());
  useEffect(() => {
    const t = setInterval(() => setNow(new Date()), 1000);
    return () => clearInterval(t);
  }, []);
  return <span className="clock">{now.toLocaleTimeString()}</span>;
}

export default function Dashboard() {
  const nav = useNavigate();
  const [me, setMe] = useState(null);
  const [loading, setLoading] = useState(true);
  const [msg, setMsg] = useState("");

  useEffect(() => {
    (async () => {
      try {
        const data = await apiMe(); // Throws 401 if no cookie/token
        setMe(data);
      } catch {
        nav("/login", { replace: true });
      } finally {
        setLoading(false);
      }
    })();
  }, [nav]);

  const onLogout = async () => {
    try { await apiLogout(); } catch {}
    nav("/login", { replace: true });
  };

  if (loading) {
    return (
      <div className="hero">
        <div className="glass" style={{ maxWidth: 420, textAlign: "center" }}>
          Loading…
        </div>
      </div>
    );
  }

  return (
    <div>
      {/* Top bar */}
      <header className="topbar">
        <div className="topbar-left">
          <strong>Secure Communication LTD</strong>
        </div>
        <div className="topbar-center">
          <Clock />
        </div>
        <div className="topbar-right">
          <span className="user-chip">👤 {me?.username || me?.email}</span>
          <button className="btn ghost" onClick={() => nav("/change-password")}>
            Change password
          </button>
          <button className="btn primary" onClick={onLogout}>
            Logout
          </button>
        </div>
      </header>

      {/* Main content placeholder */}
      <main className="page">
        <div className="glass" style={{ maxWidth: 820 }}>
          <h2 style={{ marginTop: 0 }}>Welcome{me?.username ? `, ${me.username}` : ""}!</h2>
          <p className="tagline">
            This is your dashboard. Next steps: customer search & add-customer screens.
          </p>

          {msg && (
            <div className="note">{msg}</div>
          )}
        </div>
      </main>
    </div>
  );
}
