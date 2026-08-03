import { Link, useLocation, useNavigate } from "react-router-dom";
import { useAuth } from "../context/AuthContext";
import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";
import { COLORS, CONTENT_WIDTH, secondaryButtonStyle } from "../ui";
import LanguageSwitcher from "./LanguageSwitcher";

export default function Header() {
  const { user, logout } = useAuth();
  const { lang } = useLanguage();
  const tx = translations[lang];
  const navigate = useNavigate();
  const location = useLocation();
  const authenticated = user?.userRole === "user";

  const handleLogout = async () => {
    await logout();
    navigate("/", { replace: true });
  };

  return (
    <header style={{ position: "sticky", top: 0, zIndex: 20, borderBottom: `1px solid ${COLORS.faint}`, background: "rgba(9,11,10,0.86)", backdropFilter: "blur(16px)" }}>
      <div style={{ maxWidth: CONTENT_WIDTH, margin: "0 auto", minHeight: 62, padding: "0 18px", display: "flex", gap: 14, alignItems: "center", justifyContent: "space-between", boxSizing: "border-box" }}>
        <Link to="/" style={{ display: "flex", alignItems: "center", gap: 9, color: COLORS.text, textDecoration: "none", fontWeight: 850, letterSpacing: "-0.03em" }}>
          <span aria-hidden style={{ width: 28, height: 28, borderRadius: "50%", background: `radial-gradient(circle, ${COLORS.background} 0 13%, ${COLORS.green} 14% 24%, #153c23 25% 35%, ${COLORS.green} 36% 42%, #153c23 43%)`, boxShadow: `0 0 24px ${COLORS.greenDark}` }} />
          <span>{tx.brand}</span>
        </Link>
        <nav style={{ display: "flex", flexWrap: "wrap", justifyContent: "flex-end", gap: 9, alignItems: "center" }}>
          {authenticated && location.pathname !== "/dashboard" && <Link to="/dashboard" style={{ ...secondaryButtonStyle, textDecoration: "none" }}>{tx.navDashboard}</Link>}
          <LanguageSwitcher />
          {authenticated ? (
            <button onClick={handleLogout} style={secondaryButtonStyle}>{tx.logOut}</button>
          ) : (
            <Link to="/auth" style={{ ...secondaryButtonStyle, textDecoration: "none" }}>{tx.signIn}</Link>
          )}
        </nav>
      </div>
    </header>
  );
}
