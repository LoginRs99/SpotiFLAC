type EventHandler = (...args: any[]) => void;

const handlers = new Map<string, Set<EventHandler>>();

export function EventsOn(eventName: string, callback: EventHandler): void {
  const bucket = handlers.get(eventName) ?? new Set<EventHandler>();
  bucket.add(callback);
  handlers.set(eventName, bucket);
}

export function EventsOff(eventName: string): void {
  handlers.delete(eventName);
}

export function EventsEmit(eventName: string, data?: unknown): void {
  handlers.get(eventName)?.forEach((handler) => handler(data));
}

export function BrowserOpenURL(url: string): void {
  if (url) {
    window.open(url, "_blank", "noopener,noreferrer");
  }
}

export function WindowMinimise(): void {}

export function WindowToggleMaximise(): void {}

export function Quit(): void {}

export function OnFileDrop(_callback: EventHandler, _useDropTarget = false): void {}

export function OnFileDropOff(): void {}
