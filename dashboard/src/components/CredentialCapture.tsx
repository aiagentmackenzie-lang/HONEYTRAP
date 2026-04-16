import React from 'react';
import { Eye, User, Lock, Globe } from 'lucide-react';

interface Credential {
  id: string;
  session_id: string;
  username: string;
  password: string;
  source_ip: string;
  service: string;
  timestamp: string;
}

const MOCK_CREDS: Credential[] = [
  { id: '1', session_id: 's-001', username: 'admin', password: 'admin123', source_ip: '103.224.1.1', service: 'HTTP+', timestamp: new Date().toISOString() },
  { id: '2', session_id: 's-002', username: 'root', password: 'toor', source_ip: '185.220.101.1', service: 'SSH+', timestamp: new Date(Date.now() - 3600000).toISOString() },
  { id: '3', session_id: 's-003', username: 'administrator', password: 'P@ssw0rd', source_ip: '45.33.32.1', service: 'HTTP+', timestamp: new Date(Date.now() - 7200000).toISOString() },
  { id: '4', session_id: 's-004', username: 'ubuntu', password: 'ubuntu', source_ip: '210.140.1.1', service: 'SSH+', timestamp: new Date(Date.now() - 10800000).toISOString() },
];

const CredentialCapture: React.FC<{ credentials?: Credential[] }> = ({ credentials = MOCK_CREDS }) => {
  return (
    <div className="card">
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-semibold flex items-center gap-2">
          <Eye className="w-5 h-5 text-honeytrap-yellow" />
          Captured Credentials
        </h3>
        <span className="px-2 py-0.5 rounded-full bg-honeytrap-yellow/20 text-honeytrap-yellow text-xs font-mono">
          {credentials.length} captured
        </span>
      </div>

      <div className="space-y-2 max-h-64 overflow-y-auto">
        {credentials.map((cred) => (
          <div
            key={cred.id}
            className="p-3 rounded-lg bg-honeytrap-bg border border-honeytrap-border/50"
          >
            <div className="flex items-center justify-between mb-2">
              <div className="flex items-center gap-2">
                <User className="w-3.5 h-3.5 text-honeytrap-blue" />
                <span className="font-mono text-sm">{cred.username}</span>
              </div>
              <span className="text-xs text-honeytrap-muted">
                {new Date(cred.timestamp).toLocaleTimeString()}
              </span>
            </div>
            <div className="flex items-center gap-2 mb-2">
              <Lock className="w-3.5 h-3.5 text-honeytrap-red" />
              <span className="font-mono text-sm text-honeytrap-red">{cred.password}</span>
            </div>
            <div className="flex items-center gap-2 text-xs text-honeytrap-muted">
              <Globe className="w-3 h-3" />
              <span className="font-mono">{cred.source_ip}</span>
              <span className="px-1.5 py-0.5 rounded bg-honeytrap-blue/20 text-honeytrap-blue text-[10px]">
                {cred.service}
              </span>
            </div>
          </div>
        ))}
      </div>

      {credentials.length === 0 && (
        <div className="text-center py-6 text-honeytrap-muted">
          <Eye className="w-8 h-8 mx-auto mb-2 opacity-50" />
          <p>No credentials captured yet</p>
        </div>
      )}
    </div>
  );
};

export default CredentialCapture;