import { useEffect, useState } from "react";

export const THEME_COLORS = [
  { name: "Red", value: "red", primary: "#F0614B", sidebar: "#F3CBC4" },
  { name: "Orange", value: "orange", primary: "#E9A23F", sidebar: "#F3DFC4" },
  { name: "Green", value: "green", primary: "#88BA42", sidebar: "#DFF3C4" },
  { name: "Emerald", value: "emerald", primary: "#38AB51", sidebar: "#C4F3CE" },
  { name: "Teal", value: "teal", primary: "#38AB8D", sidebar: "#C4F3E6" },
  { name: "Cyan", value: "cyan", primary: "#45BBDE", sidebar: "#C4E8F3" },
  { name: "Blue", value: "blue", primary: "#456BDE", sidebar: "#C4D0F3" },
  { name: "Purple", value: "purple", primary: "#865EEA", sidebar: "#D2C4F3" },
  { name: "Pink", value: "pink", primary: "#D55EF3", sidebar: "#E9C4F3" },
  { name: "Rose", value: "rose", primary: "#F655B8", sidebar: "#F3C4E1" },
  { name: "Gray", value: "gray", primary: "#3F3F46", sidebar: "#EAEAEA" },
] as const;

const STORAGE_KEY = "theme-accent-color";
const DEFAULT_COLOR = "orange";

export function useAccentColor() {
  const [accentColor, setAccentColorState] = useState<string>(() => {
    if (typeof window === "undefined") return DEFAULT_COLOR;
    return localStorage.getItem(STORAGE_KEY) || DEFAULT_COLOR;
  });

  useEffect(() => {
    const selectedColor = THEME_COLORS.find((c) => c.value === accentColor);
    if (!selectedColor) return;

    const root = document.documentElement;

    // Update CSS variables with hex colors
    root.style.setProperty("--sidebar", selectedColor.sidebar);
    root.style.setProperty("--primary", selectedColor.primary);

    // Store in localStorage
    localStorage.setItem(STORAGE_KEY, accentColor);
  }, [accentColor]);

  const setAccentColor = (color: string) => {
    const colorExists = THEME_COLORS.find((c) => c.value === color);
    if (colorExists) {
      setAccentColorState(color);
    }
  };

  return {
    accentColor,
    setAccentColor,
    availableColors: THEME_COLORS,
  };
}
