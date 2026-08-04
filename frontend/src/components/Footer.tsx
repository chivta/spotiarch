import { useLanguage } from "../context/LanguageContext";
import { translations } from "../i18n";
import { COLORS } from "../ui";

export default function Footer() {
  const { lang } = useLanguage();
  return <footer style={{ textAlign: "center", padding: "30px 18px", color: COLORS.muted, fontSize: 12 }}>{translations[lang].footer}</footer>;
}
