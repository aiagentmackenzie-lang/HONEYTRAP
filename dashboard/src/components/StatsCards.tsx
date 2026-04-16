import React from 'react';
import { Activity, Users, AlertTriangle, Key } from 'lucide-react';

interface StatsCardsProps {
  totalSessions?: number;
  activeNow?: number;
  alertsToday?: number;
  tokensTriggered?: number;
}

const StatsCards: React.FC<StatsCardsProps> = ({
  totalSessions = 0,
  activeNow = 0,
  alertsToday = 0,
  tokensTriggered = 0,
}) => {
  const cards = [
    {
      label: 'Total Sessions',
      value: totalSessions,
      icon: Activity,
      color: 'text-honeytrap-blue',
      glow: 'shadow-[0_0_20px_rgba(50,130,184,0.15)]',
    },
    {
      label: 'Active Now',
      value: activeNow,
      icon: Users,
      color: 'text-honeytrap-green',
      glow: 'glow-green',
    },
    {
      label: 'Alerts Today',
      value: alertsToday,
      icon: AlertTriangle,
      color: 'text-honeytrap-red',
      glow: 'glow-red',
    },
    {
      label: 'Tokens Triggered',
      value: tokensTriggered,
      icon: Key,
      color: 'text-honeytrap-yellow',
      glow: 'shadow-[0_0_20px_rgba(245,197,24,0.15)]',
    },
  ];

  return (
    <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
      {cards.map((card) => (
        <div
          key={card.label}
          className={`card flex items-center gap-4 ${card.glow}`}
        >
          <div className={`p-3 rounded-lg bg-honeytrap-bg ${card.color}`}>
            <card.icon className="w-6 h-6" />
          </div>
          <div>
            <p className="text-honeytrap-muted text-xs uppercase tracking-wider">
              {card.label}
            </p>
            <p className={`text-2xl font-bold font-mono ${card.color}`}>
              {card.value.toLocaleString()}
            </p>
          </div>
        </div>
      ))}
    </div>
  );
};

export default StatsCards;