import React from 'react';
import { X, Clock, Globe, Server, AlertTriangle } from 'lucide-react';

interface Event {
  id: string;
  session_id: string;
  type: string;
  data: string;
  timestamp: string;
}

interface SessionDetailProps {
  sessionId: string;
  ip: string;
  service: string;
  startedAt: string;
  endedAt?: string;
  status: string;
  riskScore?: number;
  events: Event[];
  onClose: () => void;
}

const eventTypeColors: Record<string, string> = {
  command: 'text-honeytrap-green',
  login: 'text-honeytrap-blue',
  error: 'text-honeytrap-red',
  data: 'text-honeytrap-yellow',
  connect: 'text-honeytrap-green',
  disconnect: 'text-honeytrap-muted',
};

const SessionDetail: React.FC<SessionDetailProps> = ({
  sessionId,
  ip,
  service,
  startedAt,
  endedAt,
  status,
  riskScore,
  events,
  onClose,
}) => {
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
      <div className="card w-full max-w-2xl max-h-[80vh] overflow-hidden animate-slide-in">
        <div className="flex items-center justify-between p-4 border-b border-honeytrap-border">
          <h3 className="text-lg font-semibold">Session Detail</h3>
          <button onClick={onClose} className="p-1 hover:bg-honeytrap-border rounded">
            <X className="w-5 h-5" />
          </button>
        </div>

        <div className="p-4 grid grid-cols-2 gap-3 text-sm border-b border-honeytrap-border">
          <div className="flex items-center gap-2">
            <Globe className="w-4 h-4 text-honeytrap-blue" />
            <span className="text-honeytrap-muted">IP:</span>
            <span className="font-mono">{ip}</span>
          </div>
          <div className="flex items-center gap-2">
            <Server className="w-4 h-4 text-honeytrap-green" />
            <span className="text-honeytrap-muted">Service:</span>
            <span className="font-mono">{service}</span>
          </div>
          <div className="flex items-center gap-2">
            <Clock className="w-4 h-4 text-honeytrap-yellow" />
            <span className="text-honeytrap-muted">Start:</span>
            <span className="font-mono">{new Date(startedAt).toLocaleString()}</span>
          </div>
          <div className="flex items-center gap-2">
            <AlertTriangle className="w-4 h-4 text-honeytrap-red" />
            <span className="text-honeytrap-muted">Risk:</span>
            <span className="font-mono">{riskScore?.toFixed(2) ?? 'N/A'}</span>
          </div>
        </div>

        <div className="p-4 overflow-y-auto max-h-[50vh]">
          <p className="text-xs text-honeytrap-muted uppercase mb-2 tracking-wider">
            Event Timeline ({events.length})
          </p>
          <div className="space-y-2">
            {events.map((event, i) => (
              <div
                key={event.id}
                className="flex gap-3 text-sm animate-slide-in"
                style={{ animationDelay: `${i * 30}ms` }}
              >
                <div className="flex flex-col items-center">
                  <div className={`w-2 h-2 rounded-full mt-1.5 ${
                    eventTypeColors[event.type] ? 'bg-current ' + eventTypeColors[event.type] : 'bg-honeytrap-muted'
                  }`} />
                  {i < events.length - 1 && (
                    <div className="w-px h-full bg-honeytrap-border" />
                  )}
                </div>
                <div className="pb-3 flex-1">
                  <div className="flex items-center justify-between">
                    <span className={`font-mono text-xs ${eventTypeColors[event.type] || 'text-honeytrap-muted'}`}>
                      {event.type}
                    </span>
                    <span className="text-xs text-honeytrap-muted font-mono">
                      {new Date(event.timestamp).toLocaleTimeString()}
                    </span>
                  </div>
                  <p className="font-mono text-xs mt-0.5 text-honeytrap-text/80 break-all">
                    {event.data}
                  </p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </div>
    </div>
  );
};

export default SessionDetail;