/**
 * Logger utility for conditional logging based on environment
 */
export class Logger {
  private static isDevelopment = import.meta.env.DEV;

  static log(...args: any[]): void {
    if (this.isDevelopment) {
      console.log(...args);
    }
  }

  static error(...args: any[]): void {
    if (this.isDevelopment) {
      console.error(...args);
    }
  }

  static warn(...args: any[]): void {
    if (this.isDevelopment) {
      console.warn(...args);
    }
  }

  static debug(...args: any[]): void {
    if (this.isDevelopment) {
      console.log(...args);
    }
  }

  static group(label: string): void {
    if (this.isDevelopment) {
      console.group(label);
    }
  }

  static groupEnd(): void {
    if (this.isDevelopment) {
      console.groupEnd();
    }
  }

  static table(data: any): void {
    if (this.isDevelopment) {
      console.table(data);
    }
  }

  static info(...args: any[]): void {
    if (this.isDevelopment) {
      console.info(...args);
    }
  }
}
