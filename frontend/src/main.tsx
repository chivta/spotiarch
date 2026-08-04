import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import App from "./App";
import { COLORS } from "./ui";

const ROOT_ID = "app";
const container = document.getElementById(ROOT_ID);
if (!container) throw new Error("Root element unavailable");

Object.assign(document.body.style, {
  margin: "0",
  minWidth: "320px",
  background: COLORS.background,
  fontFamily: "Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif",
});

createRoot(container).render(<StrictMode><App /></StrictMode>);
