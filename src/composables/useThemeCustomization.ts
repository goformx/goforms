import { ref, watch, type Ref } from "vue";

export interface ThemeConfig {
  primary: string;
  primaryForeground: string;
  secondary: string;
  secondaryForeground: string;
  accent: string;
  accentForeground: string;
  destructive: string;
  destructiveForeground: string;
  muted: string;
  mutedForeground: string;
  border: string;
  radius: string;
  fontFamily: string;
}

export interface ThemePreset {
  name: string;
  description: string;
  config: Partial<ThemeConfig>;
}

export interface UseThemeCustomizationReturn {
  theme: Ref<ThemeConfig>;
  isLoading: Ref<boolean>;
  error: Ref<string | null>;
  loadTheme: () => Promise<void>;
  saveTheme: () => Promise<void>;
  applyTheme: (config: Partial<ThemeConfig>) => void;
  applyPreset: (presetName: string) => void;
  resetTheme: () => void;
  exportTheme: () => string;
  importTheme: (json: string) => void;
  presets: ThemePreset[];
}

// Default theme configuration
const DEFAULT_THEME: ThemeConfig = {
  primary: "222.2 47.4% 11.2%",
  primaryForeground: "210 40% 98%",
  secondary: "210 40% 96.1%",
  secondaryForeground: "222.2 47.4% 11.2%",
  accent: "210 40% 96.1%",
  accentForeground: "222.2 47.4% 11.2%",
  destructive: "0 84.2% 60.2%",
  destructiveForeground: "210 40% 98%",
  muted: "210 40% 96.1%",
  mutedForeground: "215.4 16.3% 46.9%",
  border: "214.3 31.8% 91.4%",
  radius: "0.5rem",
  fontFamily: "Inter, system-ui, sans-serif",
};

// Theme presets inspired by modern design systems
const THEME_PRESETS: ThemePreset[] = [
  {
    name: "Linear",
    description: "Clean, minimal design inspired by Linear",
    config: {
      primary: "0 0% 9%",
      primaryForeground: "0 0% 98%",
      secondary: "240 5% 96%",
      secondaryForeground: "240 6% 10%",
      accent: "240 5% 96%",
      accentForeground: "240 6% 10%",
      border: "240 6% 90%",
      radius: "0.375rem",
    },
  },
  {
    name: "Stripe",
    description: "Professional design inspired by Stripe",
    config: {
      primary: "240 5.9% 10%",
      primaryForeground: "0 0% 98%",
      secondary: "240 4.8% 95.9%",
      secondaryForeground: "240 5.9% 10%",
      accent: "635 91% 25%",
      accentForeground: "0 0% 100%",
      border: "240 5.9% 90%",
      radius: "0.5rem",
    },
  },
  {
    name: "Notion",
    description: "Warm, friendly design inspired by Notion",
    config: {
      primary: "0 0% 9%",
      primaryForeground: "0 0% 98%",
      secondary: "37 15% 95%",
      secondaryForeground: "0 0% 9%",
      accent: "37 91% 55%",
      accentForeground: "0 0% 100%",
      border: "37 10% 88%",
      radius: "0.25rem",
    },
  },
  {
    name: "Vercel",
    description: "Modern, high-contrast design inspired by Vercel",
    config: {
      primary: "0 0% 0%",
      primaryForeground: "0 0% 100%",
      secondary: "0 0% 96.1%",
      secondaryForeground: "0 0% 9%",
      accent: "0 0% 96.1%",
      accentForeground: "0 0% 9%",
      border: "0 0% 89.8%",
      radius: "0.5rem",
    },
  },
];

/**
 * Composable for managing form theme customization
 *
 * @param formId Optional form ID for persistence
 * @returns Theme customization interface
 *
 * @example
 * ```ts
 * const {
 *   theme,
 *   applyTheme,
 *   applyPreset,
 *   saveTheme
 * } = useThemeCustomization('form-123');
 *
 * // Apply a custom color
 * applyTheme({ primary: '220 90% 50%' });
 *
 * // Apply a preset
 * applyPreset('Linear');
 *
 * // Save to server
 * await saveTheme();
 * ```
 */
export function useThemeCustomization(
  formId?: string,
): UseThemeCustomizationReturn {
  const theme = ref<ThemeConfig>({ ...DEFAULT_THEME });
  const isLoading = ref(false);
  const error = ref<string | null>(null);

  /**
   * Apply theme CSS variables to the document
   */
  const applyThemeVariables = (): void => {
    const root = document.documentElement;

    // Apply color variables
    root.style.setProperty("--primary", theme.value.primary);
    root.style.setProperty(
      "--primary-foreground",
      theme.value.primaryForeground,
    );
    root.style.setProperty("--secondary", theme.value.secondary);
    root.style.setProperty(
      "--secondary-foreground",
      theme.value.secondaryForeground,
    );
    root.style.setProperty("--accent", theme.value.accent);
    root.style.setProperty("--accent-foreground", theme.value.accentForeground);
    root.style.setProperty("--destructive", theme.value.destructive);
    root.style.setProperty(
      "--destructive-foreground",
      theme.value.destructiveForeground,
    );
    root.style.setProperty("--muted", theme.value.muted);
    root.style.setProperty("--muted-foreground", theme.value.mutedForeground);
    root.style.setProperty("--border", theme.value.border);

    // Apply radius
    root.style.setProperty("--radius", theme.value.radius);

    // Apply font family
    root.style.setProperty("--font-sans", theme.value.fontFamily);
  };

  /**
   * Load theme from server or localStorage
   */
  const loadTheme = async (): Promise<void> => {
    if (!formId) return;

    isLoading.value = true;
    error.value = null;

    try {
      // Try to fetch from API first
      const response = await fetch(`/api/v1/forms/${formId}/theme`);

      if (response.ok) {
        const data = (await response.json()) as { theme: ThemeConfig };
        theme.value = { ...DEFAULT_THEME, ...data.theme };
      } else {
        // Fallback to localStorage
        const savedTheme = localStorage.getItem(`form-theme-${formId}`);
        if (savedTheme) {
          theme.value = {
            ...DEFAULT_THEME,
            ...JSON.parse(savedTheme),
          } as ThemeConfig;
        }
      }

      applyThemeVariables();
    } catch (err) {
      error.value = err instanceof Error ? err.message : "Failed to load theme";
      console.error("Failed to load theme:", err);
    } finally {
      isLoading.value = false;
    }
  };

  /**
   * Save theme to server and localStorage
   */
  const saveTheme = async (): Promise<void> => {
    if (!formId) return;

    isLoading.value = true;
    error.value = null;

    try {
      // Save to localStorage first (immediate feedback)
      localStorage.setItem(`form-theme-${formId}`, JSON.stringify(theme.value));

      // Try to save to API
      const response = await fetch(`/api/v1/forms/${formId}/theme`, {
        method: "PUT",
        headers: {
          "Content-Type": "application/json",
          "X-Requested-With": "XMLHttpRequest",
        },
        body: JSON.stringify({ theme: theme.value }),
      });

      if (!response.ok) {
        throw new Error("Failed to save theme to server");
      }
    } catch (err) {
      error.value = err instanceof Error ? err.message : "Failed to save theme";
      console.error("Failed to save theme:", err);
    } finally {
      isLoading.value = false;
    }
  };

  /**
   * Apply custom theme configuration
   */
  const applyTheme = (config: Partial<ThemeConfig>): void => {
    theme.value = { ...theme.value, ...config };
    applyThemeVariables();
  };

  /**
   * Apply a theme preset
   */
  const applyPreset = (presetName: string): void => {
    const preset = THEME_PRESETS.find((p) => p.name === presetName);
    if (!preset) {
      console.warn(`Theme preset "${presetName}" not found`);
      return;
    }

    applyTheme(preset.config);
  };

  /**
   * Reset theme to default
   */
  const resetTheme = (): void => {
    theme.value = { ...DEFAULT_THEME };
    applyThemeVariables();
  };

  /**
   * Export theme as JSON string
   */
  const exportTheme = (): string => {
    return JSON.stringify(theme.value, null, 2);
  };

  /**
   * Import theme from JSON string
   */
  const importTheme = (json: string): void => {
    try {
      const imported = JSON.parse(json) as Partial<ThemeConfig>;
      applyTheme(imported);
    } catch (err) {
      error.value = "Invalid theme JSON";
      console.error("Failed to import theme:", err);
    }
  };

  // Watch theme changes and auto-apply
  watch(
    theme,
    () => {
      applyThemeVariables();
    },
    { deep: true },
  );

  // Apply initial theme
  applyThemeVariables();

  return {
    theme,
    isLoading,
    error,
    loadTheme,
    saveTheme,
    applyTheme,
    applyPreset,
    resetTheme,
    exportTheme,
    importTheme,
    presets: THEME_PRESETS,
  };
}
