import React, { useRef, useEffect } from 'react';
import * as d3 from 'd3';

interface ServiceData {
  service: string;
  count: number;
}

const MOCK_DATA: ServiceData[] = [
  { service: 'SSH', count: 156 },
  { service: 'SSH+', count: 89 },
  { service: 'HTTP', count: 134 },
  { service: 'HTTP+', count: 67 },
  { service: 'FTP', count: 43 },
  { service: 'Redis', count: 28 },
  { service: 'UDP', count: 15 },
];

const SERVICE_COLORS: Record<string, string> = {
  SSH: '#4ecca3',
  'SSH+': '#3282b8',
  HTTP: '#e84545',
  'HTTP+': '#f5c518',
  FTP: '#9b59b6',
  Redis: '#e67e22',
  UDP: '#1abc9c',
};

const ServiceChart: React.FC<{ data?: ServiceData[] }> = ({ data = MOCK_DATA }) => {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    if (!svgRef.current) return;

    const margin = { top: 10, right: 20, bottom: 30, left: 40 };
    const width = 500 - margin.left - margin.right;
    const height = 250 - margin.top - margin.bottom;

    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();
    svg.attr('viewBox', `0 0 ${width + margin.left + margin.right} ${height + margin.top + margin.bottom}`);

    const g = svg.append('g').attr('transform', `translate(${margin.left},${margin.top})`);

    const x = d3.scaleBand().domain(data.map(d => d.service)).range([0, width]).padding(0.3);
    const y = d3.scaleLinear().domain([0, d3.max(data, d => d.count)!]).range([height, 0]);

    // Grid lines
    g.append('g')
      .selectAll('line')
      .data(y.ticks(5))
      .join('line')
      .attr('x1', 0)
      .attr('x2', width)
      .attr('y1', d => y(d))
      .attr('y2', d => y(d))
      .attr('stroke', '#2a2a4a')
      .attr('stroke-width', 0.5);

    // Bars
    g.selectAll('.bar')
      .data(data)
      .join('rect')
      .attr('class', 'bar')
      .attr('x', d => x(d.service)!)
      .attr('y', d => y(d.count))
      .attr('width', x.bandwidth())
      .attr('height', d => height - y(d.count))
      .attr('fill', d => SERVICE_COLORS[d.service] || '#4ecca3')
      .attr('rx', 3)
      .attr('opacity', 0.85);

    // X axis
    g.append('g')
      .attr('transform', `translate(0,${height})`)
      .call(d3.axisBottom(x).tickSize(0))
      .selectAll('text')
      .attr('fill', '#6c6c8a')
      .attr('font-size', '10px')
      .attr('font-family', 'JetBrains Mono, monospace');
    g.select('.domain').attr('stroke', '#2a2a4a');

    // Y axis
    g.append('g')
      .call(d3.axisLeft(y).ticks(5).tickSize(0).tickFormat(d3.format('d')))
      .selectAll('text')
      .attr('fill', '#6c6c8a')
      .attr('font-size', '10px')
      .attr('font-family', 'JetBrains Mono, monospace');
    g.select('.domain').attr('stroke', 'transparent');

    // Value labels
    g.selectAll('.val')
      .data(data)
      .join('text')
      .attr('x', d => x(d.service)! + x.bandwidth() / 2)
      .attr('y', d => y(d.count) - 5)
      .attr('text-anchor', 'middle')
      .attr('fill', '#e0e0e0')
      .attr('font-size', '10px')
      .attr('font-family', 'JetBrains Mono, monospace')
      .text(d => d.count);
  }, [data]);

  return (
    <div className="card">
      <h3 className="text-lg font-semibold mb-3">Attacks by Service</h3>
      <svg ref={svgRef} className="w-full h-auto" />
    </div>
  );
};

export default ServiceChart;