// src/pages/Dashboard.jsx
import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiMe, apiLogout, apiSearchCustomers } from "../lib/api";

function Clock() {
  const [now, setNow] = useState(new Date());
  useEffect(() => {
    const t = setInterval(() => setNow(new Date()), 1000);
    return () => clearInterval(t);
  }, []);
  return <span className="clock">{now.toLocaleTimeString()}</span>;
}

function useDebounce(value, delay = 300) {
  const [v, setV] = useState(value);
  useEffect(() => {
    const t = setTimeout(() => setV(value), delay);
    return () => clearTimeout(t);
  }, [value, delay]);
  return v;
}

export default function Dashboard() {
  const nav = useNavigate();

  // session
  const [me, setMe] = useState(null);
  const [loadingMe, setLoadingMe] = useState(true);

  // search & paging
  const [q, setQ] = useState("");
  const debouncedQ = useDebounce(q, 300);
  const [page, setPage] = useState(1);
  const [size] = useState(10);
  const [rows, setRows] = useState([]);
  const [total, setTotal] = useState(0);
  const [loadingSearch, setLoadingSearch] = useState(false);
  const [searchErr, setSearchErr] = useState("");

  // load current user (guard)
  useEffect(() => {
    (async () => {
      try {
        const data = await apiMe();
        setMe(data);
      } catch {
        nav("/login", { replace: true });
      } finally {
        setLoadingMe(false);
      }
    })();
  }, [nav]);

  // search only when input >= 2 chars
  useEffect(() => {
    if (loadingMe) return;

    const term = debouncedQ.trim();
    if (term.length < 2) {
      // clear view if query too short
      setRows([]);
      setTotal(0);
      setSearchErr("");
      setLoadingSearch(false);
      return;
    }

    (async () => {
      setLoadingSearch(true);
      setSearchErr("");
      try {
        const data = await apiSearchCustomers({ q: term, page, size });
        setRows(Array.isArray(data.items) ? data.items : []);
        setTotal(typeof data.total === "number" ? data.total : 0);
      } catch (err) {
        setRows([]);
        setTotal(0);
        setSearchErr(err?.message || "Search failed");
      } finally {
        setLoadingSearch(false);
      }
    })();
  }, [debouncedQ, page, size, loadingMe]);

  const onLogout = async () => {
    try { await apiLogout(); } catch {}
    nav("/login", { replace: true });
  };

  if (loadingMe) {
    return (
      <div className="hero">
        <div className="glass" style={{ maxWidth: 420, textAlign: "center" }}>
          Loading…
        </div>
      </div>
    );
  }

  const pages = Math.max(1, Math.ceil(total / size));
  const showHelper = q.trim().length < 2;

  return (
    <div>
      {/* Top bar */}
      <header className="topbar">
        <div className="topbar-left"><strong>Secure Communication LTD</strong></div>
        <div className="topbar-center"><Clock /></div>
        <div className="topbar-right">
          <span className="user-chip">👤 {me?.username || me?.email}</span>
          <button className="btn ghost" onClick={() => nav("/change-password")}>Change password</button>
          <button className="btn primary" onClick={() => nav("/customers/new")}>New Customer</button>
          <button className="btn primary" onClick={onLogout}>Logout</button>
        </div>
      </header>

      {/* Main */}
      <main className="page">
        <div className="glass" style={{ width: "min(92vw, 1024px)" }}>
          <h2 style={{ marginTop: 0 }}>Welcome{me?.username ? `, ${me.username}` : ""}!</h2>

          {/* Search bar */}
          <div style={{ display: "flex", gap: 12, margin: "12px 0 18px" }}>
            <input
              className="input"
              placeholder="Search customers by name, email, or notes… (min. 2 chars)"
              value={q}
              onChange={(e) => { setQ(e.target.value); setPage(1); }}
              style={{ flex: 1 }}
            />
          </div>

          {/* Helper / Error */}
          {searchErr && (
            <div className="note" style={{ marginBottom: 12 }}>
              {searchErr}
            </div>
          )}
          {showHelper && !searchErr && (
            <div className="note" style={{ marginBottom: 12 }}>
              Type at least 2 characters to search.
            </div>
          )}

          {/* Results table */}
          <div style={{ overflowX: "auto" }}>
            <table style={{ width: "100%", borderCollapse: "collapse" }}>
              <thead>
                <tr style={{ opacity: .85 }}>
                  <th style={th}>ID</th>
                  <th style={th}>Name</th>
                  <th style={th}>Email</th>
                  <th style={th}>Phone</th>
                  <th style={th}>Notes</th>
                  <th style={th}>Created</th>
                </tr>
              </thead>
              <tbody>
                {showHelper ? (
                  <tr><td colSpan={6} style={td}>Start typing to search…</td></tr>
                ) : loadingSearch ? (
                  <tr><td colSpan={6} style={td}>Loading…</td></tr>
                ) : rows.length === 0 ? (
                  <tr><td colSpan={6} style={td}>No results</td></tr>
                ) : rows.map(r => (
                  <tr key={r.id}>
                    <td style={td}>{r.id}</td>
                    <td style={td}>{r.name}</td>
                    <td style={td}>{r.email}</td>
                    <td style={td}>{r.phone || "-"}</td>
                    {}
                    <td style={td} title={r.notes || ""}>{r.notes || "-"}</td>
                    <td style={td}>{new Date(r.created_at).toLocaleString()}</td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          {/* Pagination */}
          {!showHelper && (
            <div style={{ display: "flex", justifyContent: "flex-end", gap: 8, marginTop: 12 }}>
              <button className="btn ghost" disabled={page <= 1} onClick={() => setPage(p => p - 1)}>Prev</button>
              <div style={{ alignSelf: "center", opacity: .8 }}>Page {page} / {pages}</div>
              <button className="btn ghost" disabled={page >= pages} onClick={() => setPage(p => p + 1)}>Next</button>
            </div>
          )}
        </div>
      </main>
    </div>
  );
}

const th = { textAlign: "left", padding: "10px 8px", borderBottom: "1px solid rgba(255,255,255,0.12)" };
const td = { padding: "10px 8px", borderBottom: "1px solid rgba(255,255,255,0.06)" };
