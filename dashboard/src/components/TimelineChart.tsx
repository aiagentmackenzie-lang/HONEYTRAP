import React, { useRef, useEffect } from 'react';
import * as d3 from 'd3';

interface TimeData {
  hour: string;
  count: number;
}

const MOCK_DATA: TimeData[] = Array.from({ length: 24 }, (_, i) => ({
  hour: `${String(i).padStart(2, '0')}:00`,
  count: Math.floor(Math.random() * 50) + 5,
}));

const TimelineChart: React.FC<{ data?: TimeData[] }> = ({ data = MOCK_DATA }) => {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    if (!svgRef.current) return;

    const margin = { top: 10, right: 20, bottom: 30, left: 40 };
    const width = 600 - margin.left - margin.right;
    const height = 200 - margin.top - margin.bottom;

    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();
    svg.attr('viewBox', `0 0 ${width + margin.left + margin.right} ${height + margin.top + margin.bottom}`);

    const g = svg.append('g').attr('transform', `translate(${margin.left},${margin.top})`);

    const x = d3.scalePoint().domain(data.map(d => d.hour)).range([0, width]);
    const y = d3.scaleLinear().domain([0, d3.max(data, d => d.count)!]).range([height, 0]);

    // Grid lines
    g.append('g')
      .selectAll('line')
      .data(y.ticks(4))
      .join('line')
      .attr('x1', 0)
      .attr('x2', width)
      .attr('y1', d => y(d))
      .attr('y2', d => y(d))
      .attr('stroke', '#2a2a4a')
      .attr('stroke-width', 0.5);

    // Area
    const area = d3.area<TimeData>()
      .x(d => x(d.hour)!)
      .y0(height)
      .y1(d => y(d.count))
      .curve(d3.curveCatmullRom);

    const gradient = svg.append('defs')
      .append('linearGradient')
      .attr('id', 'areaGradient')
      .attr('x1', '0').attr('y1', '0')
      .attr('x2', '0').attr('y2', '1');
    gradient.append('stop').attr('offset', '0%').attr('stop-color', '#4ecca3').attr('stop-opacity', 0.3);
    gradient.append('stop').attr('offset', '100%').attr('stop-color', '#4ecca3').attr('stop-opacity', 0.02);

    g.append('path')
      .datum(data)
      .attr('d', area)
      .attr('fill', 'url(#areaGradient)');

    // Line
    const line = d3.line<TimeData>()
      .x(d => x(d.hour)!)
      .y(d => y(d.count))
      .curve(d3.curveCatmullRom);

    g.append('path')
      .datum(data)
      .attr('d', line)
      .attr('fill', 'none')
      .attr('stroke', '#4ecca3')
      .attr('stroke-width', 2);

    // X axis
    g.append('g')
      .attr('transform', `translate(0,${height})`)
      .call(d3.axisBottom(x).tickValues(data.filter((_, i) => i % 4 === 0).map(d => d.hour)).tickSize(0))
      .selectAll('text')
      .attr('fill', '#6c6c8a')
      .attr('font-size', '9px')
      .attr('font-family', 'JetBrains Mono, monospace');
    g.select('.domain').attr('stroke', '#2a2a4a');

    // Y axis
    g.append('g')
      .call(d3.axisLeft(y).ticks(4).tickSize(0).tickFormat(d3.format('d')))
      .selectAll('text')
      .attr('fill', '#6c6c8a')
      .attr('font-size', '9px')
      .attr('font-family', 'JetBrains Mono, monospace');
    g.select('.domain').attr('stroke', 'transparent');
  }, [data]);

  return (
    <div className="card">
      <h3 className="text-lg font-semibold mb-3">Attacks (24h)</h3>
      <svg ref={svgRef} className="w-full h-auto" />
    </div>
  );
};

export default TimelineChart;