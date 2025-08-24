import React from "react";
import "./App.css";

export default function App() {
  const goSignUp = () => (window.location.href = "/register"); // Can be replaced with a navigator later
  const goLogin  = () => (window.location.href = "/login");

  return (
    <div className="hero">
      <div className="glass">
        <h1 className="brand">Communication_LTD</h1>
        <p className="tagline">Secure communication. Seamless experience.</p>

        <div className="actions">
          <button className="btn primary" onClick={goSignUp}>Create Account</button>
          <button className="btn ghost"   onClick={goLogin}>Sign In</button>
        </div>

        <div className="footer-note">
          <span className="dot" /> Secure Version
        </div>
      </div>

      <div className="background-blur blur-1" />
      <div className="background-blur blur-2" />
      <div className="background-blur blur-3" />
    </div>
  );
}
