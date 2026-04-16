import React from 'react';
import AttackerMap from '../components/AttackerMap';
import ServiceChart from '../components/ServiceChart';
import TimelineChart from '../components/TimelineChart';

const AnalyticsPage: React.FC = () => {
  return (
    <div className="space-y-6 max-w-7xl mx-auto">
      <div>
        <h1 className="text-2xl font-bold">Analytics</h1>
        <p className="text-honeytrap-muted text-sm mt-1">
          Attack patterns, geolocation, and service targeting
        </p>
      </div>

      <AttackerMap />

      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        <ServiceChart />
        <TimelineChart />
      </div>
    </div>
  );
};

export default AnalyticsPage;