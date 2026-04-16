import React from 'react';
import { NavLink } from 'react-router-dom';
import {
  LayoutDashboard,
  MonitorPlay,
  KeyRound,
  BarChart3,
  Settings,
  Shield,
} from 'lucide-react';

const navItems = [
  { to: '/', icon: LayoutDashboard, label: 'Dashboard' },
  { to: '/sessions', icon: MonitorPlay, label: 'Sessions' },
  { to: '/tokens', icon: KeyRound, label: 'Tokens' },
  { to: '/analytics', icon: BarChart3, label: 'Analytics' },
  { to: '/settings', icon: Settings, label: 'Settings' },
];

const Sidebar: React.FC = () => {
  return (
    <aside className="w-56 bg-honeytrap-card border-r border-honeytrap-border flex flex-col h-screen sticky top-0">
      <div className="p-4 border-b border-honeytrap-border">
        <div className="flex items-center gap-2">
          <Shield className="w-6 h-6 text-honeytrap-green" />
          <div>
            <h1 className="text-lg font-bold tracking-tight">HONEYTRAP</h1>
            <p className="text-[10px] text-honeytrap-muted uppercase tracking-widest">
              Deception Framework
            </p>
          </div>
        </div>
      </div>

      <nav className="flex-1 p-3 space-y-1">
        {navItems.map((item) => (
          <NavLink
            key={item.to}
            to={item.to}
            end={item.to === '/'}
            className={({ isActive }) =>
              `flex items-center gap-3 px-3 py-2.5 rounded-lg text-sm transition-all ${
                isActive
                  ? 'bg-honeytrap-green/10 text-honeytrap-green border border-honeytrap-green/20 glow-green'
                  : 'text-honeytrap-muted hover:text-honeytrap-text hover:bg-honeytrap-border/30'
              }`
            }
          >
            <item.icon className="w-4 h-4" />
            <span>{item.label}</span>
          </NavLink>
        ))}
      </nav>

      <div className="p-4 border-t border-honeytrap-border">
        <div className="flex items-center gap-2">
          <div className="pulse-dot" />
          <span className="text-xs text-honeytrap-muted">System Online</span>
        </div>
      </div>
    </aside>
  );
};

export default Sidebar;