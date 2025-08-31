import React from "react";
import { Routes, Route, useNavigate } from "react-router-dom";
import Register from "./pages/Register.jsx";
import Login from "./pages/Login.jsx";
import Forgot from "./pages/Forgot";
import Reset from "./pages/Reset";

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

const Placeholder = ({ title }) => (
  <div className="hero"><div className="glass"><h2>{title}</h2></div></div>
);

export default function App() {
  return (
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/register" element={<Register />} />
      <Route path="/login" element={<Login/>} />
      <Route path="/forgot" element={<Forgot />} />
      <Route path="/reset" element={<Reset />} />
    </Routes>
  );
}
