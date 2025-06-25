/**
 * Test setup file for Vitest
 * Configures the testing environment with DOM utilities and global mocks
 */

import { vi } from "vitest";

// Mock DOM APIs that might not be available in jsdom
Object.defineProperty(window, "location", {
  value: {
    href: "http://localhost:5173",
    origin: "http://localhost:5173",
    pathname: "/",
    search: "",
    hash: "",
  },
  writable: true,
});

// Mock console methods to reduce noise in tests
global.console = {
  ...console,
  debug: vi.fn(),
  log: vi.fn(),
  warn: vi.fn(),
  error: vi.fn(),
};

// Mock fetch for API testing
global.fetch = vi.fn();

// Mock localStorage
const localStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
Object.defineProperty(window, "localStorage", {
  value: localStorageMock,
});

// Mock sessionStorage
const sessionStorageMock = {
  getItem: vi.fn(),
  setItem: vi.fn(),
  removeItem: vi.fn(),
  clear: vi.fn(),
};
Object.defineProperty(window, "sessionStorage", {
  value: sessionStorageMock,
});

// Mock FormData
global.FormData = class FormData {
  private data = new Map<string, any>();

  append(key: string, value: any): void {
    this.data.set(key, value);
  }

  get(key: string): any {
    return this.data.get(key);
  }

  has(key: string): boolean {
    return this.data.has(key);
  }

  delete(key: string): void {
    this.data.delete(key);
  }

  entries(): IterableIterator<[string, any]> {
    return this.data.entries();
  }

  forEach(callback: (value: any, key: string) => void): void {
    this.data.forEach(callback);
  }

  keys(): IterableIterator<string> {
    return this.data.keys();
  }

  values(): IterableIterator<any> {
    return this.data.values();
  }
} as any;

// Mock Headers
global.Headers = class Headers {
  private headers = new Map<string, string>();

  constructor(init?: Record<string, string>) {
    if (init) {
      Object.entries(init).forEach(([key, value]) => {
        this.headers.set(key.toLowerCase(), value);
      });
    }
  }

  append(name: string, value: string): void {
    this.headers.set(name.toLowerCase(), value);
  }

  delete(name: string): void {
    this.headers.delete(name.toLowerCase());
  }

  get(name: string): string | null {
    return this.headers.get(name.toLowerCase()) || null;
  }

  has(name: string): boolean {
    return this.headers.has(name.toLowerCase());
  }

  set(name: string, value: string): void {
    this.headers.set(name.toLowerCase(), value);
  }

  entries(): IterableIterator<[string, string]> {
    return this.headers.entries();
  }

  forEach(callback: (value: string, key: string) => void): void {
    this.headers.forEach(callback);
  }

  keys(): IterableIterator<string> {
    return this.headers.keys();
  }

  values(): IterableIterator<string> {
    return this.headers.values();
  }
} as any;

// Mock Response
global.Response = class Response {
  public ok: boolean;
  public status: number;
  public statusText: string;
  public headers: Headers;
  private body: any;

  constructor(body?: any, init?: ResponseInit) {
    this.body = body;
    this.ok = (init?.status || 200) >= 200 && (init?.status || 200) < 300;
    this.status = init?.status || 200;
    this.statusText = init?.statusText || "";
    this.headers = new Headers(init?.headers);
  }

  async json(): Promise<any> {
    return this.body;
  }

  async text(): Promise<string> {
    return typeof this.body === "string"
      ? this.body
      : JSON.stringify(this.body);
  }

  async formData(): Promise<FormData> {
    return new FormData();
  }
} as any;

// Mock Request
global.Request = class Request {
  public url: string;
  public method: string;
  public headers: Headers;
  public body: any;

  constructor(input: string | Request, init?: RequestInit) {
    this.url = typeof input === "string" ? input : input.url;
    this.method = init?.method || "GET";
    this.headers = new Headers(init?.headers);
    this.body = init?.body;
  }
} as any;
