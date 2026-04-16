import React from 'react';
import { Wifi, WifiOff, Shield, Server } from 'lucide-react';

interface Service {
  name: string;
  port: number;
  protocol: 'TCP' | 'UDP';
  status: 'running' | 'stopped' | 'error';
  connections: number;
}

const MOCK_SERVICES: Service[] = [
  { name: 'SSH', port: 2222, protocol: 'TCP', status: 'running', connections: 12 },
  { name: 'SSH+', port: 2223, protocol: 'TCP', status: 'running', connections: 5 },
  { name: 'HTTP', port: 8080, protocol: 'TCP', status: 'running', connections: 23 },
  { name: 'HTTP+', port: 8081, protocol: 'TCP', status: 'running', connections: 8 },
  { name: 'FTP', port: 2121, protocol: 'TCP', status: 'running', connections: 3 },
  { name: 'Redis', port: 6379, protocol: 'TCP', status: 'running', connections: 2 },
  { name: 'UDP', port: 5353, protocol: 'UDP', status: 'running', connections: 1 },
];

const ServiceStatus: React.FC<{ services?: Service[] }> = ({ services = MOCK_SERVICES }) => {
  return (
    <div className="card">
      <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
        <Shield className="w-5 h-5 text-honeytrap-green" />
        Service Status
      </h3>
      <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-3">
        {services.map((svc) => (
          <div
            key={svc.name}
            className={`p-3 rounded-lg border ${
              svc.status === 'running'
                ? 'border-honeytrap-green/30 bg-honeytrap-green/5'
                : 'border-honeytrap-red/30 bg-honeytrap-red/5'
            }`}
          >
            <div className="flex items-center justify-between mb-2">
              <span className="font-mono text-sm font-semibold">{svc.name}</span>
              {svc.status === 'running' ? (
                <Wifi className="w-4 h-4 text-honeytrap-green" />
              ) : (
                <WifiOff className="w-4 h-4 text-honeytrap-red" />
              )}
            </div>
            <div className="flex items-center justify-between text-xs text-honeytrap-muted">
              <span className="font-mono">:{svc.port}</span>
              <span>{svc.protocol}</span>
            </div>
            <div className="mt-2 text-xs">
              <span className="text-honeytrap-muted">Connections: </span>
              <span className="font-mono text-honeytrap-text">{svc.connections}</span>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default ServiceStatus;