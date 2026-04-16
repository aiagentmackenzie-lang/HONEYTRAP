import React, { useState } from 'react';
import TokenList from '../components/TokenList';
import TokenAlerts from '../components/TokenAlerts';
import { useApi, apiPost, apiDelete } from '../hooks/useApi';
import { Plus, Key } from 'lucide-react';

interface Token {
  id: string;
  kind: string;
  value: string;
  status: 'active' | 'triggered' | 'deactivated';
  access_count: number;
  created_at: string;
  last_accessed?: string;
}

const MOCK_TOKENS: Token[] = [
  { id: 't-1', kind: 'api-key', value: 'sk-proj-4ecca3deadbeef1234567890abcdef', status: 'active', access_count: 0, created_at: new Date().toISOString() },
  { id: 't-2', kind: 'aws-creds', value: 'AKIADEADBEEF12345678', status: 'triggered', access_count: 3, created_at: new Date(Date.now() - 86400000).toISOString(), last_accessed: new Date().toISOString() },
  { id: 't-3', kind: 'db-url', value: 'postgresql://admin:password@db.internal:5432/prod', status: 'active', access_count: 0, created_at: new Date(Date.now() - 172800000).toISOString() },
  { id: 't-4', kind: 'document', value: 'https://internal.corp/secrets/backup.key', status: 'deactivated', access_count: 1, created_at: new Date(Date.now() - 259200000).toISOString() },
];

const TOKEN_KINDS = ['api-key', 'aws-creds', 'db-url', 'document'];

const TokensPage: React.FC = () => {
  const { data: apiTokens, refetch } = useApi<Token[]>('/tokens');
  const [showCreate, setShowCreate] = useState(false);
  const [newKind, setNewKind] = useState('api-key');

  const tokens = (apiTokens && apiTokens.length > 0) ? apiTokens : MOCK_TOKENS;

  const handleCreate = async () => {
    try {
      await apiPost('/tokens', { kind: newKind });
      refetch();
      setShowCreate(false);
    } catch (err) {
      console.error('Failed to create token:', err);
    }
  };

  const handleDeactivate = async (id: string) => {
    try {
      await apiDelete(`/tokens/${id}`);
      refetch();
    } catch (err) {
      console.error('Failed to deactivate token:', err);
    }
  };

  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold">Honeytokens</h1>
          <p className="text-honeytrap-muted text-sm mt-1">
            Deception assets and tracking beacons
          </p>
        </div>
        <button
          onClick={() => setShowCreate(!showCreate)}
          className="flex items-center gap-2 px-4 py-2 rounded-lg bg-honeytrap-green/20 text-honeytrap-green border border-honeytrap-green/30 hover:bg-honeytrap-green/30 transition-colors text-sm"
        >
          <Plus className="w-4 h-4" />
          Create Token
        </button>
      </div>

      {showCreate && (
        <div className="card animate-slide-in">
          <h3 className="font-semibold mb-3 flex items-center gap-2">
            <Key className="w-4 h-4 text-honeytrap-yellow" />
            New Honeytoken
          </h3>
          <div className="flex gap-3 items-end">
            <div className="flex-1">
              <label className="block text-xs text-honeytrap-muted mb-1">Kind</label>
              <select
                value={newKind}
                onChange={(e) => setNewKind(e.target.value)}
                className="w-full px-3 py-2 rounded-lg bg-honeytrap-bg border border-honeytrap-border text-sm font-mono"
              >
                {TOKEN_KINDS.map((k) => (
                  <option key={k} value={k}>{k}</option>
                ))}
              </select>
            </div>
            <button
              onClick={handleCreate}
              className="px-4 py-2 rounded-lg bg-honeytrap-green text-honeytrap-bg font-semibold text-sm hover:bg-honeytrap-green/80 transition-colors"
            >
              Generate
            </button>
            <button
              onClick={() => setShowCreate(false)}
              className="px-4 py-2 rounded-lg bg-honeytrap-border text-honeytrap-muted text-sm hover:text-honeytrap-text transition-colors"
            >
              Cancel
            </button>
          </div>
        </div>
      )}

      <TokenList tokens={tokens} onDeactivate={handleDeactivate} />

      <TokenAlerts
        alerts={[
          { id: '1', token_id: 't-2', token_kind: 'aws-creds', accessor_ip: '185.220.101.1', timestamp: new Date().toISOString(), service: 'Redis' },
          { id: '2', token_id: 't-2', token_kind: 'aws-creds', accessor_ip: '103.224.1.1', timestamp: new Date(Date.now() - 1800000).toISOString(), service: 'SSH' },
          { id: '3', token_id: 't-4', token_kind: 'document', accessor_ip: '45.33.32.1', timestamp: new Date(Date.now() - 86400000).toISOString(), service: 'HTTP+' },
        ]}
      />
    </div>
  );
};

export default TokensPage;