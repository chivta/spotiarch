import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { createWatch, getPending, issueVerificationToken } from "../api";
import { ApiError } from "../api/client";
import ErrorNotice from "../components/ErrorNotice";
import Layout from "../components/Layout";
import PlaylistPreviewCard from "../components/PlaylistPreviewCard";
import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";
import type { PendingResponse } from "../types/models";
import { cardStyle, COLORS, CONTENT_WIDTH, primaryButtonStyle, secondaryButtonStyle } from "../ui";

export default function SetupPage() {
  const [pending, setPending] = useState<PendingResponse | null>(null);
  const [error, setError] = useState<unknown>(null);
  const [checking, setChecking] = useState(false);
  const [copied, setCopied] = useState(false);
  const { lang } = useLanguage();
  const tx = translations[lang];
  const navigate = useNavigate();

  useEffect(() => {
    let active = true;
    getPending()
      .then(next => next.step === "selected" ? issueVerificationToken() : next)
      .then(next => { if (active) setPending(next); })
      .catch(nextError => {
        if (!active) return;
        if (nextError instanceof ApiError && ["NO_PENDING_SELECTION", "PENDING_EXPIRED"].includes(nextError.code)) navigate("/", { replace: true });
        else setError(nextError);
      });
    return () => { active = false; };
  }, [navigate]);

  const copyToken = async () => {
    if (!pending?.verification_token) return;
    await navigator.clipboard.writeText(pending.verification_token);
    setCopied(true);
  };

  const handleCheck = async () => {
    setChecking(true);
    setError(null);
    try {
      const watch = await createWatch();
      navigate(`/archive/${watch.id}`, { replace: true, state: { verified: true } });
    } catch (nextError) {
      setError(nextError);
      setChecking(false);
    }
  };

  return (
    <Layout>
      <div style={{ maxWidth: CONTENT_WIDTH, margin: "0 auto", padding: "56px 18px" }}>
        {!pending && !error && <div style={{ color: COLORS.muted }}>{tx.loading}</div>}
        {Boolean(error) && <div style={{ marginBottom: 18 }}><ErrorNotice error={error} /></div>}
        {pending && (
          <>
            <div style={{ color: COLORS.green, fontSize: 12, fontWeight: 800, textTransform: "uppercase", letterSpacing: "0.13em", marginBottom: 12 }}>{tx.verificationEyebrow}</div>
            <h1 style={{ margin: "0 0 12px", fontSize: "clamp(2rem, 5vw, 3.6rem)", letterSpacing: "-0.05em" }}>{tx.verificationTitle}</h1>
            <p style={{ color: COLORS.muted, lineHeight: 1.65, maxWidth: 720, marginBottom: 10 }}>{tx.verificationBody}</p>
            <p style={{ color: COLORS.green, fontSize: 13, marginBottom: 26 }}>{tx.verificationSelectionSaved}</p>
            <PlaylistPreviewCard playlist={pending.playlist} />
            <section style={{ ...cardStyle, padding: "clamp(20px, 4vw, 32px)", marginTop: 18 }}>
              <div style={{ color: COLORS.muted, fontSize: 12, fontWeight: 700, textTransform: "uppercase", letterSpacing: "0.1em", marginBottom: 9 }}>{tx.verificationTokenLabel}</div>
              <code style={{ display: "block", overflowWrap: "anywhere", color: COLORS.green, fontSize: "clamp(1.45rem, 5vw, 2.5rem)", fontWeight: 800, letterSpacing: "0.04em", marginBottom: 20 }}>{pending.verification_token}</code>
              {error instanceof ApiError && error.code === "VERIFICATION_TOKEN_NOT_FOUND" && <p style={{ color: "#ffca80", fontSize: 14 }}>{tx.verificationRetry}</p>}
              <div style={{ display: "flex", flexWrap: "wrap", gap: 10 }}>
                <button type="button" onClick={copyToken} style={secondaryButtonStyle}>{copied ? tx.copiedToken : tx.copyToken}</button>
                <a href={pending.playlist.spotify_url} target="_blank" rel="noreferrer" style={{ ...secondaryButtonStyle, textDecoration: "none" }}>{tx.openInSpotify}</a>
                <button type="button" onClick={handleCheck} disabled={checking} style={{ ...primaryButtonStyle, opacity: checking ? 0.65 : 1 }}>{checking ? tx.checkingNow : tx.checkNow}</button>
              </div>
            </section>
          </>
        )}
      </div>
    </Layout>
  );
}
