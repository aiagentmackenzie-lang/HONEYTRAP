import React, { useEffect, useRef } from 'react';
import { useWebSocket } from '../hooks/useWebSocket';

interface EventEntry {
  id: string;
  type: string;
  data: string;
  timestamp: string;
  source_ip?: string;
  service?: string;
}

const eventTypeIcon: Record<string, string> = {
  connect: '🔗',
  disconnect: '🔌',
  command: '⌨️',
  login: '🔐',
  error: '❌',
  data: '📦',
  alert: '🚨',
};

const EventLog: React.FC = () => {
  const { lastMessage, connected } = useWebSocket({
    onMessage: () => {},
  });
  const logRef = useRef<HTMLDivElement>(null);
  const [events, setEvents] = React.useState<EventEntry[]>([]);
  const [autoScroll, setAutoScroll] = React.useState(true);

  useEffect(() => {
    if (lastMessage && lastMessage.type === 'event') {
      setEvents((prev) => [lastMessage, ...prev].slice(0, 200));
    }
  }, [lastMessage]);

  useEffect(() => {
    if (autoScroll && logRef.current) {
      logRef.current.scrollTop = 0;
    }
  }, [events, autoScroll]);

  const handleScroll = () => {
    if (!logRef.current) return;
    const { scrollTop } = logRef.current;
    setAutoScroll(scrollTop < 50);
  };

  return (
    <div className="card">
      <div className="flex items-center justify-between mb-3">
        <h3 className="text-lg font-semibold flex items-center gap-2">
          📡 Live Event Stream
        </h3>
        <div className="flex items-center gap-2">
          <div className={connected ? 'pulse-dot' : 'pulse-dot-red'} />
          <span className="text-xs text-honeytrap-muted">
            {connected ? 'Live' : 'Disconnected'}
          </span>
        </div>
      </div>

      <div
        ref={logRef}
        onScroll={handleScroll}
        className="h-64 overflow-y-auto space-y-1 font-mono text-xs bg-honeytrap-bg rounded-lg p-3 border border-honeytrap-border/50"
      >
        {events.length === 0 ? (
          <div className="text-center text-honeytrap-muted py-8">
            <p>Waiting for events...</p>
            <p className="mt-1 text-[10px]">Connect to a WebSocket feed to see live data</p>
          </div>
        ) : (
          events.map((e) => (
            <div key={e.id} className="flex gap-2 py-1 border-b border-honeytrap-border/20 animate-slide-in">
              <span className="flex-shrink-0 w-4 text-center">
                {eventTypeIcon[e.type] || '•'}
              </span>
              <span className="text-honeytrap-muted flex-shrink-0">
                {new Date(e.timestamp).toLocaleTimeString()}
              </span>
              <span className={`flex-shrink-0 px-1 rounded text-[10px] ${
                e.type === 'alert' ? 'bg-honeytrap-red/20 text-honeytrap-red' :
                e.type === 'command' ? 'bg-honeytrap-green/20 text-honeytrap-green' :
                'bg-honeytrap-blue/20 text-honeytrap-blue'
              }`}>
                {e.type}
              </span>
              <span className="truncate text-honeytrap-text/80">{e.data}</span>
            </div>
          ))
        )}
      </div>
    </div>
  );
};

export default EventLog;