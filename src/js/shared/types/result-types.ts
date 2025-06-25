/**
 * Modern Result type for functional error handling
 */
export type Result<T, E = Error> =
  | { readonly success: true; readonly data: T }
  | { readonly success: false; readonly error: E };

/**
 * Option type for nullable values
 */
export type Option<T> =
  | { readonly type: "some"; readonly value: T }
  | { readonly type: "none" };

/**
 * Async Result type for async operations
 */
export type AsyncResult<T, E = Error> = Promise<Result<T, E>>;

/**
 * Result utility functions
 */
export const Result = {
  /**
   * Create a successful result
   */
  success<T>(data: T): Result<T, never> {
    return { success: true, data };
  },

  /**
   * Create a failed result
   */
  failure<E>(error: E): Result<never, E> {
    return { success: false, error };
  },

  /**
   * Map over a successful result
   */
  map<T, U, E>(result: Result<T, E>, fn: (data: T) => U): Result<U, E> {
    return result.success ? { success: true, data: fn(result.data) } : result;
  },

  /**
   * Map over a failed result
   */
  mapError<T, E, F>(result: Result<T, E>, fn: (error: E) => F): Result<T, F> {
    return result.success
      ? result
      : { success: false, error: fn(result.error) };
  },

  /**
   * Chain operations on results
   */
  flatMap<T, U, E>(
    result: Result<T, E>,
    fn: (data: T) => Result<U, E>,
  ): Result<U, E> {
    return result.success ? fn(result.data) : result;
  },

  /**
   * Unwrap a result, throwing if it's a failure
   */
  unwrap<T, E>(result: Result<T, E>): T {
    if (result.success) {
      return result.data;
    }
    throw result.error;
  },

  /**
   * Unwrap a result with a default value for failures
   */
  unwrapOr<T, E>(result: Result<T, E>, defaultValue: T): T {
    return result.success ? result.data : defaultValue;
  },

  /**
   * Unwrap a result with a function for failures
   */
  unwrapOrElse<T, E>(result: Result<T, E>, fn: (error: E) => T): T {
    return result.success ? result.data : fn(result.error);
  },

  /**
   * Check if result is successful
   */
  isSuccess<T, E>(result: Result<T, E>): result is { success: true; data: T } {
    return result.success;
  },

  /**
   * Check if result is a failure
   */
  isFailure<T, E>(
    result: Result<T, E>,
  ): result is { success: false; error: E } {
    return !result.success;
  },
} as const;

/**
 * Option utility functions
 */
export const Option = {
  /**
   * Create a some option
   */
  some<T>(value: T): Option<T> {
    return { type: "some", value };
  },

  /**
   * Create a none option
   */
  none<T>(): Option<T> {
    return { type: "none" };
  },

  /**
   * Create an option from a nullable value
   */
  fromNullable<T>(value: T | null | undefined): Option<T> {
    return value != null ? { type: "some", value } : { type: "none" };
  },

  /**
   * Map over an option
   */
  map<T, U>(option: Option<T>, fn: (value: T) => U): Option<U> {
    return option.type === "some"
      ? { type: "some", value: fn(option.value) }
      : option;
  },

  /**
   * Flat map over an option
   */
  flatMap<T, U>(option: Option<T>, fn: (value: T) => Option<U>): Option<U> {
    return option.type === "some" ? fn(option.value) : option;
  },

  /**
   * Unwrap an option with a default value
   */
  unwrapOr<T>(option: Option<T>, defaultValue: T): T {
    return option.type === "some" ? option.value : defaultValue;
  },

  /**
   * Check if option is some
   */
  isSome<T>(option: Option<T>): option is { type: "some"; value: T } {
    return option.type === "some";
  },

  /**
   * Check if option is none
   */
  isNone<T>(option: Option<T>): option is { type: "none" } {
    return option.type === "none";
  },
} as const;

/**
 * Async Result utility functions
 */
export const AsyncResult = {
  /**
   * Create an async result from a promise
   */
  async fromPromise<T>(promise: Promise<T>): Promise<Result<T, Error>> {
    try {
      const data = await promise;
      return Result.success(data);
    } catch (error) {
      return Result.failure(
        error instanceof Error ? error : new Error(String(error)),
      );
    }
  },

  /**
   * Map over an async result
   */
  async map<T, U, E>(
    asyncResult: AsyncResult<T, E>,
    fn: (data: T) => U,
  ): Promise<Result<U, E>> {
    const result = await asyncResult;
    return Result.map(result, fn);
  },

  /**
   * Chain operations on async results
   */
  async flatMap<T, U, E>(
    asyncResult: AsyncResult<T, E>,
    fn: (data: T) => AsyncResult<U, E>,
  ): Promise<Result<U, E>> {
    const result = await asyncResult;
    return result.success ? await fn(result.data) : result;
  },
} as const;
