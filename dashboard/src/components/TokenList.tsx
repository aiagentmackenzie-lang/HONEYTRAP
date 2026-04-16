import React from 'react';
import { Key, Power, Trash2 } from 'lucide-react';

interface Token {
  id: string;
  kind: string;
  value: string;
  status: 'active' | 'triggered' | 'deactivated';
  access_count: number;
  created_at: string;
  last_accessed?: string;
}

const kindColors: Record<string, string> = {
  'api-key': 'bg-honeytrap-blue/20 text-honeytrap-blue',
  'aws-creds': 'bg-honeytrap-yellow/20 text-honeytrap-yellow',
  'db-url': 'bg-honeytrap-green/20 text-honeytrap-green',
  'document': 'bg-purple-500/20 text-purple-400',
};

const statusBadge: Record<string, string> = {
  active: 'bg-honeytrap-green/20 text-honeytrap-green',
  triggered: 'bg-honeytrap-red/20 text-honeytrap-red',
  deactivated: 'bg-honeytrap-muted/20 text-honeytrap-muted',
};

const TokenList: React.FC<{
  tokens: Token[];
  onDeactivate?: (id: string) => void;
}> = ({ tokens, onDeactivate }) => {
  return (
    <div className="card overflow-hidden">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold flex items-center gap-2">
          <Key className="w-5 h-5 text-honeytrap-yellow" />
          Honeytokens
        </h3>
        <span className="text-xs text-honeytrap-muted">
          {tokens.filter(t => t.status === 'active').length} active / {tokens.length} total
        </span>
      </div>

      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-honeytrap-muted text-xs uppercase border-b border-honeytrap-border">
              <th className="text-left py-2 px-3">Kind</th>
              <th className="text-left py-2 px-3">Value</th>
              <th className="text-left py-2 px-3">Status</th>
              <th className="text-left py-2 px-3">Accesses</th>
              <th className="text-left py-2 px-3">Created</th>
              <th className="text-left py-2 px-3"></th>
            </tr>
          </thead>
          <tbody>
            {tokens.map((t) => (
              <tr
                key={t.id}
                className="border-b border-honeytrap-border/50 hover:bg-honeytrap-border/20 transition-colors"
              >
                <td className="py-2 px-3">
                  <span className={`px-2 py-0.5 rounded text-xs font-mono ${kindColors[t.kind] || 'bg-honeytrap-border text-honeytrap-text'}`}>
                    {t.kind}
                  </span>
                </td>
                <td className="py-2 px-3 font-mono text-xs truncate max-w-[200px]">
                  {t.value.substring(0, 20)}...
                </td>
                <td className="py-2 px-3">
                  <span className={`px-2 py-0.5 rounded text-xs ${statusBadge[t.status]}`}>
                    {t.status}
                  </span>
                </td>
                <td className="py-2 px-3 font-mono text-xs">{t.access_count}</td>
                <td className="py-2 px-3 text-xs text-honeytrap-muted">
                  {new Date(t.created_at).toLocaleDateString()}
                </td>
                <td className="py-2 px-3">
                  {t.status === 'active' && onDeactivate && (
                    <button
                      onClick={() => onDeactivate(t.id)}
                      className="p-1 hover:bg-honeytrap-red/20 rounded transition-colors"
                      title="Deactivate"
                    >
                      <Power className="w-3.5 h-3.5 text-honeytrap-red" />
                    </button>
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {tokens.length === 0 && (
        <div className="text-center py-8 text-honeytrap-muted">
          <Key className="w-8 h-8 mx-auto mb-2 opacity-50" />
          <p>No honeytokens created yet</p>
        </div>
      )}
    </div>
  );
};

export default TokenList;