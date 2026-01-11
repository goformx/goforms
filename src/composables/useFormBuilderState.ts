import { ref, computed, type Ref } from "vue";

export interface FormComponent {
  key: string;
  type: string;
  label?: string;
  [key: string]: unknown;
}

export interface FormSchema {
  components?: FormComponent[];
  display?: string;
  [key: string]: unknown;
}

export interface BuilderState {
  selectedField: Ref<string | null>;
  isDirty: Ref<boolean>;
  history: Ref<FormSchema[]>;
  historyIndex: Ref<number>;
  canUndo: Ref<boolean>;
  canRedo: Ref<boolean>;
}

export interface UseFormBuilderStateReturn extends BuilderState {
  selectField: (fieldKey: string | null) => void;
  markDirty: () => void;
  markClean: () => void;
  pushHistory: (schema: FormSchema) => void;
  undo: () => FormSchema | null;
  redo: () => FormSchema | null;
  clearHistory: () => void;
  reset: () => void;
}

const MAX_HISTORY_SIZE = 50;

/**
 * Composable for managing form builder state with undo/redo support
 *
 * @param formId Form identifier for localStorage persistence
 * @returns Form builder state management interface
 *
 * @example
 * ```ts
 * const {
 *   selectedField,
 *   isDirty,
 *   canUndo,
 *   canRedo,
 *   selectField,
 *   undo,
 *   redo,
 *   pushHistory
 * } = useFormBuilderState('form-123');
 *
 * // Select a field
 * selectField('firstName');
 *
 * // Push schema to history
 * pushHistory(currentSchema);
 *
 * // Undo last change
 * const previousSchema = undo();
 * ```
 */
export function useFormBuilderState(
  formId?: string
): UseFormBuilderStateReturn {
  // State
  const selectedField = ref<string | null>(null);
  const isDirty = ref(false);
  const history = ref<FormSchema[]>([]);
  const historyIndex = ref(-1);

  // Computed
  const canUndo = computed(() => historyIndex.value > 0);
  const canRedo = computed(() => historyIndex.value < history.value.length - 1);

  // Load state from localStorage if formId provided
  if (formId) {
    try {
      const savedState = localStorage.getItem(
        `form-builder-state-${formId}`
      );
      if (savedState) {
        const parsed = JSON.parse(savedState) as {
          selectedField: string | null;
        };
        selectedField.value = parsed.selectedField;
      }
    } catch (error) {
      console.error("Failed to load form builder state:", error);
    }
  }

  /**
   * Save state to localStorage
   */
  const saveState = (): void => {
    if (!formId) return;

    try {
      const state = {
        selectedField: selectedField.value,
      };
      localStorage.setItem(
        `form-builder-state-${formId}`,
        JSON.stringify(state)
      );
    } catch (error) {
      console.error("Failed to save form builder state:", error);
    }
  };

  /**
   * Select a field in the form builder
   */
  const selectField = (fieldKey: string | null): void => {
    selectedField.value = fieldKey;
    saveState();
  };

  /**
   * Mark form as dirty (has unsaved changes)
   */
  const markDirty = (): void => {
    isDirty.value = true;
  };

  /**
   * Mark form as clean (saved)
   */
  const markClean = (): void => {
    isDirty.value = false;
  };

  /**
   * Push a new schema state to history
   */
  const pushHistory = (schema: FormSchema): void => {
    // Remove any redo history when new change is made
    if (historyIndex.value < history.value.length - 1) {
      history.value = history.value.slice(0, historyIndex.value + 1);
    }

    // Add new schema to history
    history.value.push(JSON.parse(JSON.stringify(schema)) as FormSchema);

    // Limit history size
    if (history.value.length > MAX_HISTORY_SIZE) {
      history.value.shift();
    } else {
      historyIndex.value++;
    }

    markDirty();
  };

  /**
   * Undo last change
   * @returns Previous schema or null if cannot undo
   */
  const undo = (): FormSchema | null => {
    if (!canUndo.value) return null;

    historyIndex.value--;
    return JSON.parse(
      JSON.stringify(history.value[historyIndex.value])
    ) as FormSchema;
  };

  /**
   * Redo last undone change
   * @returns Next schema or null if cannot redo
   */
  const redo = (): FormSchema | null => {
    if (!canRedo.value) return null;

    historyIndex.value++;
    return JSON.parse(
      JSON.stringify(history.value[historyIndex.value])
    ) as FormSchema;
  };

  /**
   * Clear all history
   */
  const clearHistory = (): void => {
    history.value = [];
    historyIndex.value = -1;
  };

  /**
   * Reset all state
   */
  const reset = (): void => {
    selectedField.value = null;
    isDirty.value = false;
    clearHistory();
    saveState();
  };

  return {
    // State
    selectedField,
    isDirty,
    history,
    historyIndex,
    canUndo,
    canRedo,

    // Methods
    selectField,
    markDirty,
    markClean,
    pushHistory,
    undo,
    redo,
    clearHistory,
    reset,
  };
}
