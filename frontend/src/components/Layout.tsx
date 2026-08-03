import type { ReactNode } from "react";
import { COLORS } from "../ui";
import Footer from "./Footer";
import Header from "./Header";

export default function Layout({ children }: { children: ReactNode }) {
  return (
    <div style={{ minHeight: "100vh", color: COLORS.text, background: `radial-gradient(circle at 18% 0%, ${COLORS.greenDark} 0, transparent 34%), radial-gradient(circle at 90% 25%, #172b20 0, transparent 27%), ${COLORS.background}` }}>
      <Header />
      <main>{children}</main>
      <Footer />
    </div>
  );
}
