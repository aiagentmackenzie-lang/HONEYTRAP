import React, { useRef, useEffect } from 'react';
import * as d3 from 'd3';

interface AttackPoint {
  lat: number;
  lng: number;
  ip: string;
  service: string;
  count: number;
}

const MOCK_DATA: AttackPoint[] = [
  { lat: 39.9, lng: 116.4, ip: '103.224.1.1', service: 'SSH', count: 45 },
  { lat: 55.7, lng: 37.6, ip: '185.220.101.1', service: 'HTTP', count: 32 },
  { lat: 51.5, lng: -0.1, ip: '45.33.32.1', service: 'FTP', count: 18 },
  { lat: 37.5, lng: 127.0, ip: '211.218.1.1', service: 'SSH', count: 27 },
  { lat: -23.5, lng: -46.6, ip: '177.54.1.1', service: 'Redis', count: 12 },
  { lat: 35.7, lng: 139.7, ip: '210.140.1.1', service: 'HTTP+', count: 22 },
  { lat: 48.8, lng: 2.3, ip: '51.15.1.1', service: 'SSH+', count: 15 },
  { lat: 1.3, lng: 103.8, ip: '128.199.1.1', service: 'UDP', count: 8 },
  { lat: -33.9, lng: 18.4, ip: '41.58.1.1', service: 'FTP', count: 5 },
  { lat: 52.5, lng: 13.4, ip: '195.201.1.1', service: 'SSH', count: 19 },
];

const AttackerMap: React.FC = () => {
  const svgRef = useRef<SVGSVGElement>(null);

  useEffect(() => {
    if (!svgRef.current) return;

    const width = 800;
    const height = 400;
    const svg = d3.select(svgRef.current);
    svg.selectAll('*').remove();

    svg.attr('viewBox', `0 0 ${width} ${height}`);

    const projection = d3.geoNaturalEarth1()
      .scale(width / 5.5)
      .translate([width / 2, height / 2]);

    const pathGen = d3.geoPath().projection(projection);

    // Draw graticule
    const graticule = d3.geoGraticule();
    svg.append('path')
      .datum(graticule() as any)
      .attr('d', pathGen)
      .attr('fill', 'none')
      .attr('stroke', '#1a1a2e')
      .attr('stroke-width', 0.5);

    // Draw land
    const land: any = { type: 'Sphere' };
    svg.append('path')
      .datum(land)
      .attr('d', pathGen)
      .attr('fill', '#1a1a2e')
      .attr('stroke', '#2a2a4a')
      .attr('stroke-width', 0.5);

    // Draw attack points
    MOCK_DATA.forEach((point) => {
      const [x, y] = projection([point.lng, point.lat]) as [number, number];
      if (!x || !y) return;

      // Pulse ring
      svg.append('circle')
        .attr('cx', x)
        .attr('cy', y)
        .attr('r', Math.max(3, point.count / 3))
        .attr('fill', 'none')
        .attr('stroke', '#e84545')
        .attr('stroke-width', 1)
        .attr('opacity', 0.3)
        .append('animate')
        .attr('attributeName', 'r')
        .attr('values', `${Math.max(3, point.count / 3)};${Math.max(3, point.count / 3) + 8};${Math.max(3, point.count / 3)}`)
        .attr('dur', '2s')
        .attr('repeatCount', 'indefinite');

      // Core dot
      svg.append('circle')
        .attr('cx', x)
        .attr('cy', y)
        .attr('r', Math.max(2, point.count / 6))
        .attr('fill', '#e84545')
        .attr('opacity', 0.8);

      // Label
      svg.append('text')
        .attr('x', x + 6)
        .attr('y', y + 3)
        .text(point.ip)
        .attr('fill', '#6c6c8a')
        .attr('font-size', '8px')
        .attr('font-family', 'JetBrains Mono, monospace');
    });
  }, []);

  return (
    <div className="card">
      <h3 className="text-lg font-semibold mb-3 flex items-center gap-2">
        <span className="text-honeytrap-red">◉</span>
        Attacker Geolocation
      </h3>
      <div className="overflow-hidden rounded-lg bg-honeytrap-bg">
        <svg ref={svgRef} className="w-full h-auto" />
      </div>
    </div>
  );
};

export default AttackerMap;