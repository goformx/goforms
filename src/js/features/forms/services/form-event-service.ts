import { Logger } from "@/core/logger";

type EventHandler = (data: any) => void;

export class FormEventService {
  private static instance: FormEventService;
  private readonly eventHandlers: Map<string, Set<EventHandler>>;
  private readonly sessionId: string;

  private constructor() {
    this.eventHandlers = new Map();
    this.sessionId = this.generateSessionId();
  }

  public static getInstance(): FormEventService {
    if (!FormEventService.instance) {
      FormEventService.instance = new FormEventService();
    }
    return FormEventService.instance;
  }

  private generateSessionId(): string {
    return `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }

  // Event registration
  public on(eventType: string, handler: EventHandler): void {
    if (!this.eventHandlers.has(eventType)) {
      this.eventHandlers.set(eventType, new Set());
    }
    this.eventHandlers.get(eventType)!.add(handler);
  }

  public off(eventType: string, handler: EventHandler): void {
    const handlers = this.eventHandlers.get(eventType);
    if (handlers) {
      handlers.delete(handler);
    }
  }

  // Event emission
  private emit(eventType: string, data: any): void {
    const handlers = this.eventHandlers.get(eventType);
    if (handlers) {
      handlers.forEach((handler) => {
        try {
          handler(data);
        } catch (error) {
          Logger.error(`Error in event handler for ${eventType}:`, error);
        }
      });
    }
  }

  // Form lifecycle events
  public emitFormCreated(formId: string, userId: string): void {
    this.emit("form.created", {
      formId,
      userId,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFormUpdated(formId: string, userId: string): void {
    this.emit("form.updated", {
      formId,
      userId,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFormDeleted(formId: string, userId: string): void {
    this.emit("form.deleted", {
      formId,
      userId,
      timestamp: new Date().toISOString(),
    });
  }

  // Form state events
  public emitFormLoaded(formId: string): void {
    this.emit("form.loaded", {
      formId,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFormReady(formId: string): void {
    this.emit("form.ready", {
      formId,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFormDirty(formId: string): void {
    this.emit("form.dirty", {
      formId,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFormClean(formId: string): void {
    this.emit("form.clean", {
      formId,
      timestamp: new Date().toISOString(),
    });
  }

  // Form submission events
  public emitFormSubmitted(formId: string, submissionId: string): void {
    this.emit("form.submitted", {
      formId,
      submissionId,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFormValidated(
    formId: string,
    isValid: boolean,
    errors?: any[],
  ): void {
    this.emit("form.validated", {
      formId,
      isValid,
      errors,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFormProcessed(formId: string, processingId: string): void {
    this.emit("form.processed", {
      formId,
      processingId,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFormError(formId: string, error: Error, errorType: string): void {
    this.emit("form.error", {
      formId,
      error: error.message,
      errorType,
      timestamp: new Date().toISOString(),
    });
  }

  // Field events
  public emitFieldFocused(
    formId: string,
    fieldId: string,
    fieldName: string,
  ): void {
    this.emit("form.field.focused", {
      formId,
      fieldId,
      fieldName,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFieldBlurred(
    formId: string,
    fieldId: string,
    fieldName: string,
  ): void {
    this.emit("form.field.blurred", {
      formId,
      fieldId,
      fieldName,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFieldChanged(
    formId: string,
    fieldId: string,
    fieldName: string,
    value: any,
  ): void {
    this.emit("form.field.changed", {
      formId,
      fieldId,
      fieldName,
      value,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFieldValidated(
    formId: string,
    fieldId: string,
    fieldName: string,
    isValid: boolean,
    errors?: any[],
  ): void {
    this.emit("form.field.validated", {
      formId,
      fieldId,
      fieldName,
      isValid,
      errors,
      timestamp: new Date().toISOString(),
    });
  }

  // Analytics events
  public emitFormViewed(formId: string, userId: string): void {
    this.emit("form.viewed", {
      formId,
      userId,
      sessionId: this.sessionId,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFormInteracted(
    formId: string,
    userId: string,
    interactionType: string,
  ): void {
    this.emit("form.interacted", {
      formId,
      userId,
      sessionId: this.sessionId,
      interactionType,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFormAbandoned(
    formId: string,
    userId: string,
    reason?: string,
  ): void {
    this.emit("form.abandoned", {
      formId,
      userId,
      sessionId: this.sessionId,
      reason,
      timestamp: new Date().toISOString(),
    });
  }

  public emitFormCompleted(
    formId: string,
    userId: string,
    completionTime: number,
  ): void {
    this.emit("form.completed", {
      formId,
      userId,
      sessionId: this.sessionId,
      completionTime,
      timestamp: new Date().toISOString(),
    });
  }

  // Helper method to track form abandonment
  public trackFormAbandonment(formId: string, userId: string): void {
    let lastInteraction = Date.now();
    const abandonmentTimeout = 5 * 60 * 1000; // 5 minutes

    const checkAbandonment = () => {
      const now = Date.now();
      if (now - lastInteraction > abandonmentTimeout) {
        this.emitFormAbandoned(formId, userId, "timeout");
        return;
      }
      setTimeout(checkAbandonment, 60000); // Check every minute
    };

    // Update last interaction on any form event
    this.on("form.field.changed", () => {
      lastInteraction = Date.now();
    });

    this.on("form.submitted", () => {
      lastInteraction = Date.now();
    });

    // Start checking for abandonment
    checkAbandonment();
  }
}
