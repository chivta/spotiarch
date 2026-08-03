import { useEffect, useState } from "react";
import { Link } from "react-router-dom";
import { getWatches } from "../api";
import ErrorNotice from "../components/ErrorNotice";
import Layout from "../components/Layout";
import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";
import type { WatchResponse } from "../types/models";
import { cardStyle, COLORS, CONTENT_WIDTH, linkStyle, primaryButtonStyle, secondaryButtonStyle } from "../ui";

export default function Dashboard() {
  const [watches, setWatches] = useState<WatchResponse[] | null>(null);
  const [error, setError] = useState<unknown>(null);
  const { lang } = useLanguage();
  const tx = translations[lang];
  const locale = lang === "ru" ? "ru-RU" : "en-US";

  useEffect(() => { getWatches().then(setWatches).catch(setError); }, []);

  return (
    <Layout>
      <div style={{ maxWidth: CONTENT_WIDTH, margin: "0 auto", padding: "56px 18px" }}>
        <div style={{ display: "flex", flexWrap: "wrap", gap: 18, alignItems: "end", justifyContent: "space-between", marginBottom: 30 }}>
          <div><h1 style={{ margin: "0 0 8px", fontSize: "clamp(2.2rem, 5vw, 3.8rem)", letterSpacing: "-0.055em" }}>{tx.dashboardTitle}</h1><p style={{ margin: 0, color: COLORS.muted }}>{tx.dashboardBody}</p></div>
          <Link to="/" style={{ ...primaryButtonStyle, textDecoration: "none" }}>{tx.newArchive}</Link>
        </div>
        {Boolean(error) && <ErrorNotice error={error} />}
        {!watches && !error && <div style={{ color: COLORS.muted }}>{tx.loading}</div>}
        {watches?.length === 0 && <section style={{ ...cardStyle, textAlign: "center", padding: 42 }}><h2>{tx.emptyArchivesTitle}</h2><p style={{ color: COLORS.muted, marginBottom: 24 }}>{tx.emptyArchivesBody}</p><Link to="/" style={{ ...primaryButtonStyle, textDecoration: "none" }}>{tx.archiveNow}</Link></section>}
        <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 310px), 1fr))", gap: 16 }}>
          {watches?.map(watch => (
            <article key={watch.id} style={{ ...cardStyle, padding: 18 }}>
              <div style={{ display: "flex", gap: 15, alignItems: "center", marginBottom: 18 }}>
                <a href={watch.source.spotify_url} target="_blank" rel="noreferrer"><img src={watch.source.image_url} alt={watch.source.name} style={{ width: 76, height: 76, objectFit: "cover", borderRadius: 10, display: "block" }} /></a>
                <div style={{ minWidth: 0 }}><a href={watch.source.spotify_url} target="_blank" rel="noreferrer" style={linkStyle}><h2 style={{ margin: "0 0 6px", fontSize: 19, overflowWrap: "anywhere" }}>{watch.source.name}</h2></a><span style={{ color: COLORS.muted, fontSize: 13 }}>{tx.byOwner}{" "}<a href={watch.source.owner_url} target="_blank" rel="noreferrer" style={{ color: COLORS.text }}>{watch.source.owner_name}</a></span></div>
              </div>
              <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 10, marginBottom: 16 }}>
                <div style={{ background: COLORS.faint, borderRadius: 12, padding: 12 }}><strong style={{ display: "block", fontSize: 22 }}>{watch.archived_total}</strong><span style={{ color: COLORS.muted, fontSize: 12 }}>{tx.archivedTotal}</span></div>
                <div style={{ background: COLORS.faint, borderRadius: 12, padding: 12 }}><strong style={{ display: "block", fontSize: 22 }}>{watch.removed_total}</strong><span style={{ color: COLORS.muted, fontSize: 12 }}>{tx.removedTotal}</span></div>
              </div>
              <p style={{ color: COLORS.muted, fontSize: 12, margin: "0 0 17px" }}>{tx.lastPolled}: {watch.last_polled_at ? new Intl.DateTimeFormat(locale, { dateStyle: "medium", timeStyle: "short" }).format(new Date(watch.last_polled_at)) : tx.notPolledYet}</p>
              <Link to={`/archive/${watch.id}`} style={{ ...secondaryButtonStyle, display: "inline-block", textDecoration: "none" }}>{tx.viewArchive}</Link>
            </article>
          ))}
        </div>
      </div>
    </Layout>
  );
}
