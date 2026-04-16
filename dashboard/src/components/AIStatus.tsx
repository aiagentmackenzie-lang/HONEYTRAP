import React from 'react';
import { Cpu, HardDrive, Activity, Wifi } from 'lucide-react';
import { useApi } from '../hooks/useApi';

interface AIHealth {
  status: string;
  model: string;
  cache_size: number;
  uptime: number;
}

const AIStatus: React.FC = () => {
  const { data, loading, error } = useApi<AIHealth>('/ai/health');

  const health = data || {
    status: 'offline',
    model: 'N/A',
    cache_size: 0,
    uptime: 0,
  };

  return (
    <div className="card">
      <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
        <Cpu className="w-5 h-5 text-honeytrap-blue" />
        AI Emulator Status
      </h3>

      <div className="grid grid-cols-2 gap-3">
        <div className="p-3 rounded-lg bg-honeytrap-bg border border-honeytrap-border/50">
          <div className="flex items-center gap-2 mb-1">
            <Activity className="w-3.5 h-3.5 text-honeytrap-muted" />
            <span className="text-xs text-honeytrap-muted uppercase">Status</span>
          </div>
          <div className="flex items-center gap-2">
            <div className={health.status === 'ok' || health.status === 'healthy' ? 'pulse-dot' : 'pulse-dot-red'} />
            <span className={`font-mono text-sm ${
              health.status === 'ok' || health.status === 'healthy'
                ? 'text-honeytrap-green'
                : 'text-honeytrap-red'
            }`}>
              {loading ? '...' : health.status}
            </span>
          </div>
        </div>

        <div className="p-3 rounded-lg bg-honeytrap-bg border border-honeytrap-border/50">
          <div className="flex items-center gap-2 mb-1">
            <Cpu className="w-3.5 h-3.5 text-honeytrap-muted" />
            <span className="text-xs text-honeytrap-muted uppercase">Model</span>
          </div>
          <span className="font-mono text-sm text-honeytrap-blue">
            {loading ? '...' : health.model || 'N/A'}
          </span>
        </div>

        <div className="p-3 rounded-lg bg-honeytrap-bg border border-honeytrap-border/50">
          <div className="flex items-center gap-2 mb-1">
            <HardDrive className="w-3.5 h-3.5 text-honeytrap-muted" />
            <span className="text-xs text-honeytrap-muted uppercase">Cache</span>
          </div>
          <span className="font-mono text-sm text-honeytrap-yellow">
            {loading ? '...' : health.cache_size} entries
          </span>
        </div>

        <div className="p-3 rounded-lg bg-honeytrap-bg border border-honeytrap-border/50">
          <div className="flex items-center gap-2 mb-1">
            <Wifi className="w-3.5 h-3.5 text-honeytrap-muted" />
            <span className="text-xs text-honeytrap-muted uppercase">Uptime</span>
          </div>
          <span className="font-mono text-sm text-honeytrap-text">
            {loading ? '...' : health.uptime > 0 ? `${Math.floor(health.uptime / 60)}m` : 'N/A'}
          </span>
        </div>
      </div>

      {error && (
        <p className="text-xs text-honeytrap-red mt-2">⚠️ {error}</p>
      )}
    </div>
  );
};

export default AIStatus;