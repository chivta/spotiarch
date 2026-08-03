import type { PlaylistPreview } from "../types/models";
import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";
import { cardStyle, COLORS, linkStyle } from "../ui";

export default function PlaylistPreviewCard({ playlist, action }: { playlist: PlaylistPreview; action?: React.ReactNode }) {
  const { lang } = useLanguage();
  const tx = translations[lang];
  return (
    <section style={{ ...cardStyle, padding: 18, display: "flex", flexWrap: "wrap", gap: 18, alignItems: "center" }}>
      <a href={playlist.spotify_url} target="_blank" rel="noreferrer" style={{ flexShrink: 0 }}>
        <img src={playlist.image_url} alt={playlist.name} style={{ width: 132, height: 132, borderRadius: 12, objectFit: "cover", display: "block", background: COLORS.faint }} />
      </a>
      <div style={{ flex: "1 1 240px", minWidth: 0 }}>
        <div style={{ color: COLORS.green, fontSize: 12, fontWeight: 750, textTransform: "uppercase", letterSpacing: "0.1em", marginBottom: 7 }}>{tx.previewTitle}</div>
        <a href={playlist.spotify_url} target="_blank" rel="noreferrer" style={linkStyle}><h2 style={{ margin: "0 0 8px", fontSize: 25, overflowWrap: "anywhere" }}>{playlist.name}</h2></a>
        <div style={{ color: COLORS.muted, fontSize: 14 }}>
          {tx.byOwner}{" "}<a href={playlist.owner_url} target="_blank" rel="noreferrer" style={{ color: COLORS.text }}>{playlist.owner_name}</a>{" · "}<a href={playlist.spotify_url} target="_blank" rel="noreferrer" style={{ color: "inherit" }}>{tx.trackCount(playlist.track_count)}</a>
        </div>
      </div>
      {action && <div style={{ flex: "0 0 auto" }}>{action}</div>}
    </section>
  );
}
