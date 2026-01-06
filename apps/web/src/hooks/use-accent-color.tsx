/* eslint-disable react-refresh/only-export-components */
import { createContext, useContext, useEffect, useState } from "react";
import { useTheme } from "@/components/themes/theme-provider";
import { LOCAL_STORAGE_KEY } from "@/constants";

export const THEME_COLORS = [
  {
    name: "Gray",
    value: "gray",
    thumb: "#3F3F46",
    primary: "#f17455",
    sidebar: {
      light: "#fafafa",
      dark: "#171717",
    },
  },
  {
    name: "Red",
    value: "red",
    thumb: "#F0614B",
    primary: "#F0614B",
    sidebar: "#F3CBC4",
  },
  {
    name: "Orange",
    value: "orange",
    thumb: "#E9A23F",
    primary: "#E9A23F",
    sidebar: "#F3DFC4",
  },
  {
    name: "Green",
    value: "green",
    thumb: "#88BA42",
    primary: "#88BA42",
    sidebar: "#DFF3C4",
  },
  {
    name: "Emerald",
    value: "emerald",
    thumb: "#38AB51",
    primary: "#38AB51",
    sidebar: "#C4F3CE",
  },
  {
    name: "Teal",
    value: "teal",
    thumb: "#38AB8D",
    primary: "#38AB8D",
    sidebar: "#C4F3E6",
  },
  {
    name: "Cyan",
    value: "cyan",
    thumb: "#45BBDE",
    primary: "#45BBDE",
    sidebar: "#C4E8F3",
  },
  {
    name: "Blue",
    value: "blue",
    thumb: "#456BDE",
    primary: "#456BDE",
    sidebar: "#C4D0F3",
  },
  {
    name: "Purple",
    value: "purple",
    thumb: "#865EEA",
    primary: "#865EEA",
    sidebar: "#D2C4F3",
  },
  {
    name: "Pink",
    value: "pink",
    thumb: "#D55EF3",
    primary: "#D55EF3",
    sidebar: "#E9C4F3",
  },
  {
    name: "Rose",
    value: "rose",
    thumb: "#F655B8",
    primary: "#F655B8",
    sidebar: "#F3C4E1",
  },
] as const;

const DEFAULT_COLOR = "gray";

type AccentColorContextType = {
  accentColor: string;
  setAccentColor: (color: string) => void;
  availableColors: typeof THEME_COLORS;
};

const AccentColorContext = createContext<AccentColorContextType | undefined>(
  undefined,
);

export function AccentColorProvider({
  children,
}: {
  children: React.ReactNode;
}) {
  const { theme } = useTheme();
  const [accentColor, setAccentColorState] = useState<string>(() => {
    if (typeof window === "undefined") return DEFAULT_COLOR;
    return (
      localStorage.getItem(LOCAL_STORAGE_KEY.ACCENT_COLOR) || DEFAULT_COLOR
    );
  });

  useEffect(() => {
    const selectedColor = THEME_COLORS.find((c) => c.value === accentColor);
    if (!selectedColor) return;

    const root = document.documentElement;

    const updateColors = () => {
      // Determine if we're in dark mode
      const isDark =
        theme === "dark" ||
        (theme === "system" &&
          window.matchMedia("(prefers-color-scheme: dark)").matches);

      // Update CSS variables with hex colors
      const sidebarColor =
        typeof selectedColor.sidebar === "string"
          ? selectedColor.sidebar
          : isDark
            ? selectedColor.sidebar.dark
            : selectedColor.sidebar.light;

      // Use requestAnimationFrame to ensure this runs after theme class changes
      requestAnimationFrame(() => {
        root.style.setProperty("--sidebar", sidebarColor);
        root.style.setProperty("--primary", selectedColor.primary);
      });
    };

    // Initial update
    updateColors();

    // Listen for system theme changes when in system mode
    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)");
    const handleChange = () => {
      if (theme === "system") {
        updateColors();
      }
    };

    mediaQuery.addEventListener("change", handleChange);

    // Store in localStorage
    localStorage.setItem(LOCAL_STORAGE_KEY.ACCENT_COLOR, accentColor);

    return () => {
      mediaQuery.removeEventListener("change", handleChange);
    };
  }, [accentColor, theme]);

  const setAccentColor = (color: string) => {
    const colorExists = THEME_COLORS.find((c) => c.value === color);
    if (colorExists) {
      setAccentColorState(color);
    }
  };

  const value = {
    accentColor,
    setAccentColor,
    availableColors: THEME_COLORS,
  };

  return (
    <AccentColorContext.Provider value={value}>
      {children}
    </AccentColorContext.Provider>
  );
}

export function useAccentColor() {
  const context = useContext(AccentColorContext);
  if (context === undefined) {
    throw new Error("useAccentColor must be used within AccentColorProvider");
  }
  return context;
}
