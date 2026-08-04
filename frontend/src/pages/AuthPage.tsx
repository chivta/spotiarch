import { useState } from "react";
import type { FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { getPending, login, signup } from "../api";
import { ApiError } from "../api/client";
import ErrorNotice from "../components/ErrorNotice";
import Layout from "../components/Layout";
import { useAuth } from "../context/AuthContext";
import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";
import { cardStyle, COLORS, inputStyle, primaryButtonStyle } from "../ui";

type Mode = "login" | "signup";

export default function AuthPage() {
  const [mode, setMode] = useState<Mode>("signup");
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<unknown>(null);
  const { refresh } = useAuth();
  const { lang } = useLanguage();
  const tx = translations[lang];
  const navigate = useNavigate();

  const handleSubmit = async (event: FormEvent) => {
    event.preventDefault();
    setLoading(true);
    setError(null);
    try {
      await (mode === "login" ? login : signup)({ email, password });
      await refresh();
      try {
        await getPending();
        navigate("/setup", { replace: true });
      } catch (pendingError) {
        if (pendingError instanceof ApiError && pendingError.code === "NO_PENDING_SELECTION") {
          navigate("/dashboard", { replace: true });
        } else if (pendingError instanceof ApiError && pendingError.code === "PENDING_EXPIRED") {
          navigate("/", { replace: true });
        } else {
          throw pendingError;
        }
      }
    } catch (nextError) {
      setError(nextError);
      setLoading(false);
    }
  };

  return (
    <Layout>
      <div style={{ minHeight: "calc(100vh - 160px)", display: "grid", placeItems: "center", padding: "42px 18px" }}>
        <section style={{ ...cardStyle, width: "100%", maxWidth: 430, padding: "clamp(22px, 5vw, 34px)", boxSizing: "border-box" }}>
          <h1 style={{ margin: "0 0 10px", fontSize: 31, letterSpacing: "-0.04em" }}>{tx.authTitle}</h1>
          <p style={{ color: COLORS.muted, lineHeight: 1.55, margin: "0 0 24px" }}>{tx.authBody}</p>
          <div style={{ display: "flex", padding: 4, background: COLORS.faint, borderRadius: 999, marginBottom: 22 }}>
            {(["signup", "login"] as Mode[]).map(item => <button key={item} type="button" onClick={() => { setMode(item); setError(null); }} style={{ flex: 1, border: 0, borderRadius: 999, padding: 9, cursor: "pointer", fontWeight: 700, background: mode === item ? COLORS.green : "transparent", color: mode === item ? "#061109" : COLORS.muted }}>{item === "login" ? tx.logIn : tx.signUp}</button>)}
          </div>
          <form onSubmit={handleSubmit} style={{ display: "grid", gap: 15 }}>
            <label style={{ fontSize: 13, fontWeight: 700 }}>{tx.emailLabel}<input type="email" value={email} onChange={event => setEmail(event.target.value)} placeholder={tx.emailPlaceholder} required disabled={loading} style={{ ...inputStyle, marginTop: 7 }} /></label>
            <label style={{ fontSize: 13, fontWeight: 700 }}>{tx.passwordLabel}<input type="password" minLength={8} maxLength={72} value={password} onChange={event => setPassword(event.target.value)} placeholder={tx.passwordPlaceholder} required disabled={loading} style={{ ...inputStyle, marginTop: 7 }} /></label>
            {Boolean(error) && <ErrorNotice error={error} />}
            <button type="submit" disabled={loading} style={{ ...primaryButtonStyle, width: "100%", marginTop: 3, opacity: loading ? 0.65 : 1 }}>{loading ? (mode === "login" ? tx.loggingIn : tx.signingUp) : (mode === "login" ? tx.logIn : tx.signUp)}</button>
          </form>
        </section>
      </div>
    </Layout>
  );
}
