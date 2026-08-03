import type { CSSProperties } from "react";

export const COLORS = {
  background: "#090b0a",
  panel: "rgba(22, 27, 24, 0.88)",
  panelStrong: "#151a17",
  text: "#f5f7f5",
  muted: "#9aa69f",
  faint: "rgba(255,255,255,0.09)",
  green: "#1ed760",
  greenDark: "#0b3d20",
  danger: "#ff6b6b",
  warning: "#f2c94c",
} as const;

export const CONTENT_WIDTH = 1040;

export const cardStyle: CSSProperties = {
  background: COLORS.panel,
  border: `1px solid ${COLORS.faint}`,
  borderRadius: 20,
  boxShadow: "0 24px 70px rgba(0,0,0,0.28)",
};

export const primaryButtonStyle: CSSProperties = {
  border: 0,
  borderRadius: 999,
  padding: "12px 20px",
  background: COLORS.green,
  color: "#061109",
  fontSize: 14,
  fontWeight: 750,
  cursor: "pointer",
};

export const secondaryButtonStyle: CSSProperties = {
  border: "1px solid rgba(255,255,255,0.17)",
  borderRadius: 999,
  padding: "10px 17px",
  background: "transparent",
  color: COLORS.text,
  fontSize: 13,
  fontWeight: 650,
  cursor: "pointer",
};

export const inputStyle: CSSProperties = {
  width: "100%",
  boxSizing: "border-box",
  padding: "14px 16px",
  borderRadius: 12,
  border: "1px solid rgba(255,255,255,0.16)",
  outline: "none",
  background: "rgba(255,255,255,0.06)",
  color: COLORS.text,
  fontSize: 16,
};

export const linkStyle: CSSProperties = { color: "inherit", textDecoration: "none" };
