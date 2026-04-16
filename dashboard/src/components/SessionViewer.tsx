import React, { useState } from 'react';
import { ChevronRight, Clock, Globe, Server } from 'lucide-react';

interface Session {
  id: string;
  source_ip: string;
  service: string;
  started_at: string;
  ended_at?: string;
  status: 'active' | 'ended' | 'timeout';
  events_count: number;
  risk_score?: number;
}

interface SessionViewerProps {
  sessions: Session[];
  onSelect?: (session: Session) => void;
}

const statusColors: Record<string, string> = {
  active: 'text-honeytrap-green',
  ended: 'text-honeytrap-muted',
  timeout: 'text-honeytrap-yellow',
};

const statusDots: Record<string, string> = {
  active: 'pulse-dot',
  ended: 'w-2 h-2 rounded-full bg-honeytrap-muted',
  timeout: 'pulse-dot-red',
};

const SessionViewer: React.FC<SessionViewerProps> = ({ sessions, onSelect }) => {
  const [sortField, setSortField] = useState<'started_at' | 'service' | 'source_ip'>('started_at');

  const sorted = [...sessions].sort((a, b) => {
    if (sortField === 'started_at') return new Date(b.started_at).getTime() - new Date(a.started_at).getTime();
    return String(a[sortField]).localeCompare(String(b[sortField]));
  });

  return (
    <div className="card overflow-hidden">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold flex items-center gap-2">
          <Server className="w-5 h-5 text-honeytrap-green" />
          Sessions
        </h3>
        <div className="flex gap-2 text-xs">
          {(['started_at', 'service', 'source_ip'] as const).map((f) => (
            <button
              key={f}
              onClick={() => setSortField(f)}
              className={`px-2 py-1 rounded ${
                sortField === f
                  ? 'bg-honeytrap-green/20 text-honeytrap-green'
                  : 'text-honeytrap-muted hover:text-honeytrap-text'
              }`}
            >
              {f === 'source_ip' ? 'IP' : f.charAt(0).toUpperCase() + f.slice(1).replace('_', ' ')}
            </button>
          ))}
        </div>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-honeytrap-muted text-xs uppercase border-b border-honeytrap-border">
              <th className="text-left py-2 px-3">Status</th>
              <th className="text-left py-2 px-3">IP</th>
              <th className="text-left py-2 px-3">Service</th>
              <th className="text-left py-2 px-3">Started</th>
              <th className="text-left py-2 px-3">Events</th>
              <th className="text-left py-2 px-3">Risk</th>
              <th className="text-left py-2 px-3"></th>
            </tr>
          </thead>
          <tbody>
            {sorted.map((s) => (
              <tr
                key={s.id}
                className="border-b border-honeytrap-border/50 hover:bg-honeytrap-border/20 cursor-pointer transition-colors"
                onClick={() => onSelect?.(s)}
              >
                <td className="py-2 px-3">
                  <div className="flex items-center gap-2">
                    <div className={statusDots[s.status]} />
                    <span className={`text-xs ${statusColors[s.status]}`}>
                      {s.status}
                    </span>
                  </div>
                </td>
                <td className="py-2 px-3 font-mono text-xs">{s.source_ip}</td>
                <td className="py-2 px-3">
                  <span className="px-2 py-0.5 rounded bg-honeytrap-blue/20 text-honeytrap-blue text-xs">
                    {s.service}
                  </span>
                </td>
                <td className="py-2 px-3 text-xs text-honeytrap-muted">
                  {new Date(s.started_at).toLocaleTimeString()}
                </td>
                <td className="py-2 px-3 font-mono text-xs">{s.events_count}</td>
                <td className="py-2 px-3">
                  {s.risk_score !== undefined && (
                    <div className="flex items-center gap-1">
                      <div className="w-12 h-1.5 bg-honeytrap-bg rounded-full overflow-hidden">
                        <div
                          className={`h-full rounded-full ${
                            s.risk_score > 0.7
                              ? 'bg-honeytrap-red'
                              : s.risk_score > 0.4
                              ? 'bg-honeytrap-yellow'
                              : 'bg-honeytrap-green'
                          }`}
                          style={{ width: `${s.risk_score * 100}%` }}
                        />
                      </div>
                      <span className="text-xs font-mono">{s.risk_score.toFixed(2)}</span>
                    </div>
                  )}
                </td>
                <td className="py-2 px-3">
                  <ChevronRight className="w-4 h-4 text-honeytrap-muted" />
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {sessions.length === 0 && (
        <div className="text-center py-8 text-honeytrap-muted">
          <Clock className="w-8 h-8 mx-auto mb-2 opacity-50" />
          <p>No sessions recorded yet</p>
        </div>
      )}
    </div>
  );
};

export default SessionViewer;