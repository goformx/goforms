import { ref, computed, type Ref } from "vue";

export interface Command {
  id: string;
  label: string;
  icon?: string;
  keywords?: string[];
  handler: () => void;
  section?: string;
}

export interface UseCommandPaletteReturn {
  isOpen: Ref<boolean>;
  query: Ref<string>;
  filteredCommands: Ref<Command[]>;
  recentCommands: Ref<Command[]>;
  open: () => void;
  close: () => void;
  toggle: () => void;
  executeCommand: (commandId: string) => void;
  clearRecent: () => void;
}

const MAX_RECENT_COMMANDS = 10;

/**
 * Simple fuzzy match scoring
 */
function fuzzyMatch(search: string, target: string): number {
  const searchLower = search.toLowerCase();
  const targetLower = target.toLowerCase();

  // Exact match
  if (targetLower === searchLower) return 100;

  // Starts with
  if (targetLower.startsWith(searchLower)) return 90;

  // Contains
  if (targetLower.includes(searchLower)) return 70;

  // Fuzzy character match
  let score = 0;
  let searchIndex = 0;

  for (
    let i = 0;
    i < targetLower.length && searchIndex < searchLower.length;
    i++
  ) {
    if (targetLower[i] === searchLower[searchIndex]) {
      score += 1;
      searchIndex++;
    }
  }

  return searchIndex === searchLower.length
    ? 50 * (score / searchLower.length)
    : 0;
}

/**
 * Composable for command palette functionality with fuzzy search
 *
 * @param commands Array of available commands
 * @returns Command palette interface
 *
 * @example
 * ```ts
 * const commands = [
 *   {
 *     id: 'new-form',
 *     label: 'Create New Form',
 *     icon: 'Plus',
 *     handler: () => router.visit('/forms/new'),
 *     section: 'Actions'
 *   },
 *   {
 *     id: 'dashboard',
 *     label: 'Go to Dashboard',
 *     icon: 'Home',
 *     handler: () => router.visit('/dashboard'),
 *     section: 'Navigation'
 *   }
 * ];
 *
 * const {
 *   isOpen,
 *   query,
 *   filteredCommands,
 *   open,
 *   close,
 *   executeCommand
 * } = useCommandPalette(commands);
 * ```
 */
export function useCommandPalette(
  commands: Command[],
): UseCommandPaletteReturn {
  const isOpen = ref(false);
  const query = ref("");
  const recentCommands = ref<Command[]>([]);

  // Load recent commands from localStorage
  try {
    const saved = localStorage.getItem("command-palette-recent");
    if (saved) {
      const recentIds = JSON.parse(saved) as string[];
      recentCommands.value = recentIds
        .map((id) => commands.find((cmd) => cmd.id === id))
        .filter((cmd): cmd is Command => cmd !== undefined);
    }
  } catch (error) {
    console.error("Failed to load recent commands:", error);
  }

  /**
   * Filter and sort commands by search query
   */
  const filteredCommands = computed(() => {
    if (!query.value.trim()) {
      // Return recent commands if no query
      return recentCommands.value.length > 0
        ? recentCommands.value
        : commands.slice(0, 10);
    }

    // Score and filter commands
    const scored = commands
      .map((command) => {
        // Search in label
        let score = fuzzyMatch(query.value, command.label);

        // Boost score if matches keywords
        if (command.keywords) {
          for (const keyword of command.keywords) {
            const keywordScore = fuzzyMatch(query.value, keyword);
            if (keywordScore > score) {
              score = keywordScore;
            }
          }
        }

        return { command, score };
      })
      .filter((item) => item.score > 0)
      .sort((a, b) => b.score - a.score)
      .slice(0, 10);

    return scored.map((item) => item.command);
  });

  /**
   * Open command palette
   */
  const open = (): void => {
    isOpen.value = true;
    query.value = "";
  };

  /**
   * Close command palette
   */
  const close = (): void => {
    isOpen.value = false;
    query.value = "";
  };

  /**
   * Toggle command palette
   */
  const toggle = (): void => {
    if (isOpen.value) {
      close();
    } else {
      open();
    }
  };

  /**
   * Execute a command by ID
   */
  const executeCommand = (commandId: string): void => {
    const command = commands.find((cmd) => cmd.id === commandId);
    if (!command) {
      console.warn(`Command "${commandId}" not found`);
      return;
    }

    // Execute handler
    command.handler();

    // Add to recent commands
    recentCommands.value = [
      command,
      ...recentCommands.value.filter((cmd) => cmd.id !== commandId),
    ].slice(0, MAX_RECENT_COMMANDS);

    // Save to localStorage
    try {
      localStorage.setItem(
        "command-palette-recent",
        JSON.stringify(recentCommands.value.map((cmd) => cmd.id)),
      );
    } catch (error) {
      console.error("Failed to save recent commands:", error);
    }

    // Close palette
    close();
  };

  /**
   * Clear recent commands
   */
  const clearRecent = (): void => {
    recentCommands.value = [];
    try {
      localStorage.removeItem("command-palette-recent");
    } catch (error) {
      console.error("Failed to clear recent commands:", error);
    }
  };

  return {
    isOpen,
    query,
    filteredCommands,
    recentCommands,
    open,
    close,
    toggle,
    executeCommand,
    clearRecent,
  };
}
