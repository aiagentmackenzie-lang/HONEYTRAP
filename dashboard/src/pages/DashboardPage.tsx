import React from 'react';
import StatsCards from '../components/StatsCards';
import ServiceChart from '../components/ServiceChart';
import TimelineChart from '../components/TimelineChart';
import ServiceStatus from '../components/ServiceStatus';
import EventLog from '../components/EventLog';
import CredentialCapture from '../components/CredentialCapture';
import TokenAlerts from '../components/TokenAlerts';
import AIStatus from '../components/AIStatus';

const MOCK_TOKEN_ALERTS = [
  { id: '1', token_id: 't-1', token_kind: 'aws-creds', accessor_ip: '185.220.101.1', timestamp: new Date().toISOString(), service: 'Redis' },
  { id: '2', token_id: 't-2', token_kind: 'api-key', accessor_ip: '103.224.1.1', timestamp: new Date(Date.now() - 600000).toISOString(), service: 'HTTP+' },
];

const DashboardPage: React.FC = () => {
  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      <div>
        <h1 className="text-2xl font-bold">Dashboard</h1>
        <p className="text-honeytrap-muted text-sm mt-1">
          Real-time deception operations overview
        </p>
      </div>

      <StatsCards
        totalSessions={432}
        activeNow={7}
        alertsToday={23}
        tokensTriggered={5}
      />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ServiceChart />
        <TimelineChart />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        <div className="lg:col-span-2">
          <ServiceStatus />
        </div>
        <AIStatus />
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <EventLog />
        <TokenAlerts alerts={MOCK_TOKEN_ALERTS} />
      </div>

      <CredentialCapture />
    </div>
  );
};

export default DashboardPage;