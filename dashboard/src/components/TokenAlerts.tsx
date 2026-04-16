import React from 'react';
import { AlertTriangle, Key, Globe, Clock } from 'lucide-react';

interface TokenAlert {
  id: string;
  token_id: string;
  token_kind: string;
  accessor_ip: string;
  timestamp: string;
  service: string;
}

const TokenAlerts: React.FC<{ alerts: TokenAlert[] }> = ({ alerts }) => {
  return (
    <div className="card">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold flex items-center gap-2">
          <AlertTriangle className="w-5 h-5 text-honeytrap-red" />
          Token Alerts
        </h3>
        {alerts.length > 0 && (
          <span className="px-2 py-0.5 rounded-full bg-honeytrap-red/20 text-honeytrap-red text-xs font-mono animate-pulse">
            {alerts.length} alert{alerts.length !== 1 ? 's' : ''}
          </span>
        )}
      </div>

      <div className="space-y-2 max-h-64 overflow-y-auto">
        {alerts.map((alert, i) => (
          <div
            key={alert.id}
            className="flex items-center gap-3 p-2.5 rounded-lg bg-honeytrap-red/5 border border-honeytrap-red/20 animate-slide-in"
            style={{ animationDelay: `${i * 50}ms` }}
          >
            <div className="pulse-dot-red flex-shrink-0" />
            <div className="flex-1 min-w-0">
              <div className="flex items-center gap-2 text-xs">
                <Key className="w-3 h-3 text-honeytrap-yellow" />
                <span className="font-mono text-honeytrap-yellow">{alert.token_kind}</span>
                <span className="text-honeytrap-muted">accessed by</span>
                <span className="font-mono">{alert.accessor_ip}</span>
              </div>
              <div className="flex items-center gap-2 text-xs text-honeytrap-muted mt-0.5">
                <Globe className="w-3 h-3" />
                <span>{alert.service}</span>
                <Clock className="w-3 h-3 ml-2" />
                <span>{new Date(alert.timestamp).toLocaleTimeString()}</span>
              </div>
            </div>
          </div>
        ))}
      </div>

      {alerts.length === 0 && (
        <div className="text-center py-6 text-honeytrap-muted">
          <AlertTriangle className="w-8 h-8 mx-auto mb-2 opacity-50" />
          <p>No token alerts yet</p>
          <p className="text-xs mt-1">Alerts appear when honeytokens are accessed</p>
        </div>
      )}
    </div>
  );
};

export default TokenAlerts;