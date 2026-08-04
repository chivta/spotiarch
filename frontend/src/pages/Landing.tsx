import { useState } from "react";
import type { FormEvent } from "react";
import { useNavigate } from "react-router-dom";
import { issueVerificationToken, resolvePlaylist } from "../api";
import ErrorNotice from "../components/ErrorNotice";
import Layout from "../components/Layout";
import PlaylistPreviewCard from "../components/PlaylistPreviewCard";
import { useAuth } from "../context/AuthContext";
import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";
import type { PendingResponse } from "../types/models";
import { cardStyle, COLORS, CONTENT_WIDTH, inputStyle, primaryButtonStyle } from "../ui";

export default function Landing() {
  const [url, setUrl] = useState("");
  const [pending, setPending] = useState<PendingResponse | null>(null);
  const [error, setError] = useState<unknown>(null);
  const [loading, setLoading] = useState(false);
  const [starting, setStarting] = useState(false);
  const { user, refresh } = useAuth();
  const { lang } = useLanguage();
  const tx = translations[lang];
  const navigate = useNavigate();

  const handleResolve = async (event: FormEvent) => {
    event.preventDefault();
    setError(null);
    setLoading(true);
    try { setPending(await resolvePlaylist({ url })); }
    catch (nextError) { setError(nextError); }
    finally { setLoading(false); }
  };

  const handleStart = async () => {
    setStarting(true);
    setError(null);
    try {
      const currentUser = await refresh();
      if (currentUser.userRole !== "user") {
        navigate("/auth");
        return;
      }
      await issueVerificationToken();
      navigate("/setup");
    } catch (nextError) {
      setError(nextError);
      setStarting(false);
    }
  };

  return (
    <Layout>
      <div style={{ maxWidth: CONTENT_WIDTH, margin: "0 auto", padding: "78px 18px 46px" }}>
        <section style={{ maxWidth: 800, marginBottom: 42 }}>
          <div style={{ color: COLORS.green, fontSize: 13, fontWeight: 800, textTransform: "uppercase", letterSpacing: "0.13em", marginBottom: 14 }}>{tx.heroEyebrow}</div>
          <h1 style={{ margin: "0 0 18px", fontSize: "clamp(2.7rem, 8vw, 5.7rem)", lineHeight: 0.96, letterSpacing: "-0.065em", maxWidth: 850 }}>{tx.heroTitle}</h1>
          <p style={{ color: COLORS.muted, lineHeight: 1.7, fontSize: "clamp(1rem, 2vw, 1.2rem)", maxWidth: 670 }}>{tx.heroBody}</p>
        </section>

        <section style={{ ...cardStyle, padding: "clamp(18px, 4vw, 30px)", marginBottom: 20 }}>
          <form onSubmit={handleResolve}>
            <label htmlFor="playlist-url" style={{ display: "block", fontWeight: 700, marginBottom: 9 }}>{tx.playlistUrlLabel}</label>
            <div style={{ display: "flex", flexWrap: "wrap", gap: 10 }}>
              <input id="playlist-url" type="url" value={url} onChange={event => setUrl(event.target.value)} placeholder={tx.playlistUrlPlaceholder} required disabled={loading} style={{ ...inputStyle, flex: "1 1 410px" }} />
              <button type="submit" disabled={loading} style={{ ...primaryButtonStyle, opacity: loading ? 0.65 : 1 }}>{loading ? tx.resolvingPlaylist : tx.resolvePlaylist}</button>
            </div>
          </form>
          <p style={{ margin: "14px 0 0", color: COLORS.muted, fontSize: 13 }}>{tx.appendOnlyShort}</p>
        </section>

        {Boolean(error) && <div style={{ marginBottom: 18 }}><ErrorNotice error={error} /></div>}
        {pending && (
          <>
            <PlaylistPreviewCard playlist={pending.playlist} action={<button onClick={handleStart} disabled={starting} style={{ ...primaryButtonStyle, opacity: starting ? 0.65 : 1 }}>{starting ? tx.startingWatch : tx.startWatching}</button>} />
            {user?.userRole !== "user" && <p style={{ margin: "13px 12px 0", color: COLORS.muted, fontSize: 13 }}>{tx.pendingSafe}</p>}
          </>
        )}
      </div>
    </Layout>
  );
}
