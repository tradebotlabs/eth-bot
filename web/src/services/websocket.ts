import type { WSMessage } from '../types';

type MessageHandler = (message: WSMessage) => void;
type ConnectionHandler = () => void;

class WebSocketService {
  private ws: WebSocket | null = null;
  private url: string;
  private reconnectInterval = 3000;
  private maxReconnects = 50;
  private reconnectCount = 0;
  private messageHandlers: Map<string, MessageHandler[]> = new Map();
  private onOpenHandlers: ConnectionHandler[] = [];
  private onCloseHandlers: ConnectionHandler[] = [];
  private reconnectTimer: number | null = null;
  private intentionalClose = false;
  private pingInterval: number | null = null;

  constructor() {
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // Use the API server port directly
    const host = window.location.hostname;
    this.url = `${protocol}//${host}:8080/ws`;
  }

  connect(): void {
    if (this.ws?.readyState === WebSocket.OPEN || this.ws?.readyState === WebSocket.CONNECTING) {
      return;
    }

    this.intentionalClose = false;

    try {
      this.ws = new WebSocket(this.url);

      this.ws.onopen = () => {
        console.log('WebSocket connected');
        this.reconnectCount = 0;
        this.onOpenHandlers.forEach((handler) => handler());
        this.startPing();
      };

      this.ws.onclose = (event) => {
        console.log('WebSocket closed:', event.code, event.reason);
        this.stopPing();
        this.onCloseHandlers.forEach((handler) => handler());

        if (!this.intentionalClose) {
          this.scheduleReconnect();
        }
      };

      this.ws.onerror = (error) => {
        console.error('WebSocket error:', error);
      };

      this.ws.onmessage = (event) => {
        try {
          const message: WSMessage = JSON.parse(event.data);
          this.handleMessage(message);
        } catch (error) {
          console.error('Failed to parse WebSocket message:', error);
        }
      };
    } catch (error) {
      console.error('Failed to connect WebSocket:', error);
      this.scheduleReconnect();
    }
  }

  private startPing(): void {
    this.stopPing();
    this.pingInterval = window.setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'ping' }));
      }
    }, 30000);
  }

  private stopPing(): void {
    if (this.pingInterval) {
      window.clearInterval(this.pingInterval);
      this.pingInterval = null;
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectCount >= this.maxReconnects) {
      console.error('Max reconnection attempts reached');
      return;
    }

    if (this.reconnectTimer) {
      window.clearTimeout(this.reconnectTimer);
    }

    const delay = Math.min(this.reconnectInterval * Math.pow(1.5, this.reconnectCount), 30000);
    this.reconnectTimer = window.setTimeout(() => {
      this.reconnectCount++;
      console.log(`Reconnecting... (attempt ${this.reconnectCount})`);
      this.connect();
    }, delay);
  }

  disconnect(): void {
    this.intentionalClose = true;
    this.stopPing();

    if (this.reconnectTimer) {
      window.clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
    }

    if (this.ws) {
      this.ws.close(1000, 'Client disconnect');
      this.ws = null;
    }
  }

  private handleMessage(message: WSMessage): void {
    const handlers = this.messageHandlers.get(message.type);
    if (handlers) {
      handlers.forEach((handler) => handler(message));
    }

    // Also call wildcard handlers
    const wildcardHandlers = this.messageHandlers.get('*');
    if (wildcardHandlers) {
      wildcardHandlers.forEach((handler) => handler(message));
    }
  }

  subscribe(type: string, handler: MessageHandler): () => void {
    const handlers = this.messageHandlers.get(type) || [];
    handlers.push(handler);
    this.messageHandlers.set(type, handlers);

    // Return unsubscribe function
    return () => {
      const currentHandlers = this.messageHandlers.get(type) || [];
      const index = currentHandlers.indexOf(handler);
      if (index > -1) {
        currentHandlers.splice(index, 1);
        this.messageHandlers.set(type, currentHandlers);
      }
    };
  }

  onOpen(handler: ConnectionHandler): () => void {
    this.onOpenHandlers.push(handler);
    return () => {
      const index = this.onOpenHandlers.indexOf(handler);
      if (index > -1) {
        this.onOpenHandlers.splice(index, 1);
      }
    };
  }

  onClose(handler: ConnectionHandler): () => void {
    this.onCloseHandlers.push(handler);
    return () => {
      const index = this.onCloseHandlers.indexOf(handler);
      if (index > -1) {
        this.onCloseHandlers.splice(index, 1);
      }
    };
  }

  send(message: unknown): void {
    if (this.ws?.readyState === WebSocket.OPEN) {
      this.ws.send(JSON.stringify(message));
    } else {
      console.warn('WebSocket is not connected');
    }
  }

  isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }
}

export const wsService = new WebSocketService();
export default wsService;
