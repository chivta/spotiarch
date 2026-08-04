import { useEffect } from "react";
import { BrowserRouter, Navigate, Route, Routes } from "react-router-dom";
import type { ReactNode } from "react";
import Layout from "./components/Layout";
import { AuthProvider, useAuth } from "./context/AuthContext";
import { LanguageProvider, useLanguage } from "./context/LanguageContext";
import { translations } from "./i18n";
import ArchivePage from "./pages/ArchivePage";
import AuthPage from "./pages/AuthPage";
import Dashboard from "./pages/Dashboard";
import Landing from "./pages/Landing";
import SetupPage from "./pages/SetupPage";
import { COLORS } from "./ui";

function DocumentLanguage() {
  const { lang } = useLanguage();
  const tx = translations[lang];
  useEffect(() => {
    document.documentElement.lang = lang;
    document.title = tx.brand;
  }, [lang, tx.brand]);
  return null;
}

function ProtectedRoute({ children }: { children: ReactNode }) {
  const { user, loading } = useAuth();
  const { lang } = useLanguage();
  if (loading) return <Layout><div style={{ minHeight: "70vh", display: "grid", placeItems: "center", color: COLORS.muted }}>{translations[lang].loading}</div></Layout>;
  if (user?.userRole !== "user") return <Navigate to="/auth" replace />;
  return children;
}

function AppRoutes() {
  return <Routes>
    <Route path="/" element={<Landing />} />
    <Route path="/auth" element={<AuthPage />} />
    <Route path="/setup" element={<ProtectedRoute><SetupPage /></ProtectedRoute>} />
    <Route path="/dashboard" element={<ProtectedRoute><Dashboard /></ProtectedRoute>} />
    <Route path="/archive/:id" element={<ProtectedRoute><ArchivePage /></ProtectedRoute>} />
    <Route path="*" element={<Navigate to="/" replace />} />
  </Routes>;
}

export default function App() {
  return <LanguageProvider><AuthProvider><BrowserRouter><DocumentLanguage /><AppRoutes /></BrowserRouter></AuthProvider></LanguageProvider>;
}
