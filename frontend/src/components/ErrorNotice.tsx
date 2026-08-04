import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";
import { errorTranslationKey } from "../api/errors";

export default function ErrorNotice({ error }: { error: unknown }) {
  const { lang } = useLanguage();
  const tx = translations[lang];
  return <div role="alert" style={{ border: "1px solid rgba(255,107,107,0.35)", background: "rgba(255,107,107,0.08)", color: "#ffb0b0", borderRadius: 12, padding: "11px 14px", fontSize: 14 }}>{String(tx[errorTranslationKey(error)])}</div>;
}
