import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";
import type { Lang } from "../i18n";
import { COLORS } from "../ui";

const LANGUAGES: Lang[] = ["en", "uk"];

export default function LanguageSwitcher() {
  const { lang, setLang } = useLanguage();
  const tx = translations[lang];
  return (
    <label style={{ display: "flex", alignItems: "center", gap: 7, color: COLORS.muted, fontSize: 12 }}>
      <span>{tx.language}</span>
      <select
        value={lang}
        onChange={event => setLang(event.target.value as Lang)}
        aria-label={tx.language}
        style={{ background: "#121613", color: COLORS.text, border: "1px solid rgba(255,255,255,0.18)", borderRadius: 999, padding: "6px 9px", cursor: "pointer" }}
      >
        {LANGUAGES.map(code => <option key={code} value={code}>{code === "en" ? tx.english : tx.ukrainian}</option>)}
      </select>
    </label>
  );
}
