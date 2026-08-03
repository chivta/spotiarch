import { createContext, useCallback, useContext, useState } from "react";
import type { ReactNode } from "react";
import type { Lang } from "../i18n";

const STORAGE_KEY = "spotiarch_lang";

function detectLanguage(): Lang {
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (stored === "en" || stored === "ru") return stored;
  } catch { /* Storage may be unavailable. */ }
  return navigator.language?.startsWith("ru") ? "ru" : "en";
}

interface LanguageContextValue { lang: Lang; setLang: (lang: Lang) => void }
const LanguageContext = createContext<LanguageContextValue>({ lang: "en", setLang: () => undefined });

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [lang, setLanguage] = useState<Lang>(detectLanguage);
  const setLang = useCallback((next: Lang) => {
    try { localStorage.setItem(STORAGE_KEY, next); } catch { /* Storage may be unavailable. */ }
    setLanguage(next);
  }, []);
  return <LanguageContext.Provider value={{ lang, setLang }}>{children}</LanguageContext.Provider>;
}

export const useLanguage = () => useContext(LanguageContext);
