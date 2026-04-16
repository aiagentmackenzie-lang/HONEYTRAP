import React from 'react';
import ServiceStatus from '../components/ServiceStatus';
import AIStatus from '../components/AIStatus';
import { Settings, Server, Database, Cpu } from 'lucide-react';

const SettingsPage: React.FC = () => {
  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      <div>
        <h1 className="text-2xl font-bold">Settings</h1>
        <p className="text-honeytrap-muted text-sm mt-1">
          System configuration and service management
        </p>
      </div>

      <ServiceStatus />
      <AIStatus />

      <div className="card">
        <h3 className="text-lg font-semibold mb-4 flex items-center gap-2">
          <Settings className="w-5 h-5 text-honeytrap-muted" />
          Configuration
        </h3>
        <div className="space-y-3 text-sm">
          <div className="flex items-center justify-between p-3 rounded-lg bg-honeytrap-bg border border-honeytrap-border/50">
            <div className="flex items-center gap-3">
              <Server className="w-4 h-4 text-honeytrap-green" />
              <span>API Server</span>
            </div>
            <span className="font-mono text-honeytrap-muted">localhost:3000</span>
          </div>
          <div className="flex items-center justify-between p-3 rounded-lg bg-honeytrap-bg border border-honeytrap-border/50">
            <div className="flex items-center gap-3">
              <Database className="w-4 h-4 text-honeytrap-blue" />
              <span>PostgreSQL</span>
            </div>
            <span className="font-mono text-honeytrap-muted">localhost:5432</span>
          </div>
          <div className="flex items-center justify-between p-3 rounded-lg bg-honeytrap-bg border border-honeytrap-border/50">
            <div className="flex items-center gap-3">
              <Cpu className="w-4 h-4 text-honeytrap-yellow" />
              <span>AI Emulator</span>
            </div>
            <span className="font-mono text-honeytrap-muted">localhost:8000</span>
          </div>
        </div>
      </div>

      <div className="card border-honeytrap-red/20">
        <h3 className="text-lg font-semibold mb-3 text-honeytrap-red">Danger Zone</h3>
        <p className="text-sm text-honeytrap-muted mb-3">
          These actions affect all running honeypots and collected data.
        </p>
        <div className="flex gap-3">
          <button className="px-4 py-2 rounded-lg bg-honeytrap-red/10 text-honeytrap-red border border-honeytrap-red/30 text-sm hover:bg-honeytrap-red/20 transition-colors">
            Stop All Services
          </button>
          <button className="px-4 py-2 rounded-lg bg-honeytrap-red/10 text-honeytrap-red border border-honeytrap-red/30 text-sm hover:bg-honeytrap-red/20 transition-colors">
            Clear All Data
          </button>
        </div>
      </div>
    </div>
  );
};

export default SettingsPage;