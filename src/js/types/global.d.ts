import { Formio } from "@formio/js";

declare global {
  interface Window {
    formBuilder: any;
    Formio: typeof Formio;
  }
}

export {};
