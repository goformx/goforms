import { Logger } from "@/core/logger";
import type {
  FormId,
  ComponentKey,
  FieldName,
} from "@/shared/types/form-types";

/**
 * Immutable state update function type
 */
export type StateUpdater<T> = (prev: T) => T;

/**
 * State change listener type
 */
export type StateListener<T> = (state: T, prevState: T) => void;

/**
 * State subscription type
 */
export interface StateSubscription {
  readonly unsubscribe: () => void;
}

/**
 * Modern form state management with immutable patterns
 */
export class FormState {
  private static instance: FormState;
  private readonly state = new Map<string, unknown>();
  private readonly listeners = new Map<string, Set<StateListener<unknown>>>();
  private readonly abortController = new AbortController();

  private constructor() {}

  public static getInstance(): FormState {
    if (!FormState.instance) {
      FormState.instance = new FormState();
    }
    return FormState.instance;
  }

  /**
   * Set state with type safety
   */
  set<T>(key: string, value: T): void {
    const _prevValue = this.state.get(key);
    this.state.set(key, value);
    this.notifyListeners(key, value, _prevValue);
  }

  /**
   * Update state using an updater function
   */
  update<T>(key: string, updater: StateUpdater<T>): void {
    const currentValue = this.state.get(key) as T;
    const newValue = updater(currentValue);
    this.set(key, newValue);
  }

  /**
   * Get state with type safety
   */
  get<T>(key: string): T | undefined {
    return this.state.get(key) as T | undefined;
  }

  /**
   * Get state with default value
   */
  getOrDefault<T>(key: string, defaultValue: T): T {
    return (this.state.get(key) as T) ?? defaultValue;
  }

  /**
   * Check if state exists
   */
  has(key: string): boolean {
    return this.state.has(key);
  }

  /**
   * Delete state
   */
  delete(key: string): boolean {
    const hadKey = this.state.has(key);
    if (hadKey) {
      const _prevValue = this.state.get(key);
      this.state.delete(key);
      this.notifyListeners(key, undefined, _prevValue);
    }
    return hadKey;
  }

  /**
   * Clear all state
   */
  clear(): void {
    const keys = Array.from(this.state.keys());
    this.state.clear();
    keys.forEach((key) => {
      this.notifyListeners(key, undefined, this.state.get(key));
    });
  }

  /**
   * Subscribe to state changes
   */
  subscribe<T>(key: string, listener: StateListener<T>): StateSubscription {
    if (!this.listeners.has(key)) {
      this.listeners.set(key, new Set());
    }

    const keyListeners = this.listeners.get(key)!;
    keyListeners.add(listener as StateListener<unknown>);

    return {
      unsubscribe: () => {
        keyListeners.delete(listener as StateListener<unknown>);
        if (keyListeners.size === 0) {
          this.listeners.delete(key);
        }
      },
    };
  }

  /**
   * Subscribe to multiple state keys
   */
  subscribeToMultiple<T>(
    keys: readonly string[],
    listener: (updates: Record<string, T>) => void,
  ): StateSubscription {
    const subscriptions: StateSubscription[] = [];
    const updates = new Map<string, T>();

    keys.forEach((key) => {
      const subscription = this.subscribe<T>(key, (newValue, _prevValue) => {
        updates.set(key, newValue);
        listener(Object.fromEntries(updates));
      });
      subscriptions.push(subscription);
    });

    return {
      unsubscribe: () => {
        subscriptions.forEach((sub) => sub.unsubscribe());
      },
    };
  }

  /**
   * Get all state keys
   */
  getKeys(): readonly string[] {
    return Array.from(this.state.keys());
  }

  /**
   * Get state size
   */
  get size(): number {
    return this.state.size;
  }

  /**
   * Check if state is empty
   */
  get isEmpty(): boolean {
    return this.state.size === 0;
  }

  /**
   * Get state snapshot
   */
  getSnapshot(): ReadonlyMap<string, unknown> {
    return new Map(this.state);
  }

  /**
   * Abort all operations and cleanup
   */
  destroy(): void {
    this.abortController.abort();
    this.clear();
    this.listeners.clear();
  }

  /**
   * Get abort signal for cancellation
   */
  get abortSignal(): AbortSignal {
    return this.abortController.signal;
  }

  private notifyListeners(
    key: string,
    newValue: unknown,
    _prevValue: unknown,
  ): void {
    const keyListeners = this.listeners.get(key);
    if (keyListeners) {
      keyListeners.forEach((listener) => {
        try {
          listener(newValue, _prevValue);
        } catch (error) {
          console.error(`Error in state listener for key "${key}":`, error);
        }
      });
    }
  }

  /**
   * Type-safe form-specific methods
   */
  setFormData(formId: FormId, data: unknown): void {
    this.set(`form:${formId}`, data);
  }

  getFormData<T>(formId: FormId): T | undefined {
    return this.get<T>(`form:${formId}`);
  }

  setComponentData(
    formId: FormId,
    componentKey: ComponentKey,
    data: unknown,
  ): void {
    this.set(`component:${formId}:${componentKey}`, data);
  }

  getComponentData<T>(
    formId: FormId,
    componentKey: ComponentKey,
  ): T | undefined {
    return this.get<T>(`component:${formId}:${componentKey}`);
  }

  setFieldValue(formId: FormId, fieldName: FieldName, value: unknown): void {
    this.set(`field:${formId}:${fieldName}`, value);
  }

  getFieldValue<T>(formId: FormId, fieldName: FieldName): T | undefined {
    return this.get<T>(`field:${formId}:${fieldName}`);
  }
}

// Export a singleton instance for easy access
export const formState = FormState.getInstance();
