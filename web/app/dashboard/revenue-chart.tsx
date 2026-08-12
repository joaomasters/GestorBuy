"use client";

// Gráfico de área/linha de série única (receita diária). Sem legenda —
// série única, o título da seção já diz o que é plotado (regra da skill
// dataviz). Specs de marca: linha 2px, área a 10% de opacidade, gridlines
// hairline 1px sólidas, marcador ≥8px com anel na cor da superfície.
import { useRef, useState } from "react";

type Point = { date: string; revenue: number };

const WIDTH = 600;
const HEIGHT = 220;
const PAD = { top: 16, right: 12, bottom: 28, left: 56 };

const SERIES_COLOR = "#2a78d6"; // categorical slot 1 (azul)
const SURFACE_COLOR = "#fcfcfb";
const GRIDLINE_COLOR = "#e1e0d9";
const CROSSHAIR_COLOR = "#c3c2b7";
const MUTED_TEXT = "#898781";

const fullCurrency = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });
const axisCurrency = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
  notation: "compact",
  maximumFractionDigits: 1,
});

function niceMax(value: number): number {
  if (value <= 0) return 10;
  const magnitude = Math.pow(10, Math.floor(Math.log10(value)));
  const normalized = value / magnitude;
  const step = normalized <= 1 ? 1 : normalized <= 2 ? 2 : normalized <= 5 ? 5 : 10;
  return step * magnitude;
}

function formatDayLabel(iso: string): string {
  const [, month, day] = iso.split("-");
  return `${day}/${month}`;
}

export function RevenueChart({ data }: { data: Point[] }) {
  const svgRef = useRef<SVGSVGElement>(null);
  const [hoverIndex, setHoverIndex] = useState<number | null>(null);

  const innerWidth = WIDTH - PAD.left - PAD.right;
  const innerHeight = HEIGHT - PAD.top - PAD.bottom;
  const n = data.length;

  const max = niceMax(Math.max(...data.map((d) => d.revenue)) * 1.15);

  const xAt = (i: number) => PAD.left + (n === 1 ? innerWidth / 2 : (i / (n - 1)) * innerWidth);
  const yAt = (value: number) => PAD.top + innerHeight - (value / max) * innerHeight;

  const linePath = data
    .map((d, i) => `${i === 0 ? "M" : "L"} ${xAt(i).toFixed(2)} ${yAt(d.revenue).toFixed(2)}`)
    .join(" ");

  const baseY = (PAD.top + innerHeight).toFixed(2);
  const areaPath =
    `M ${xAt(0).toFixed(2)} ${baseY} ` +
    data.map((d, i) => `L ${xAt(i).toFixed(2)} ${yAt(d.revenue).toFixed(2)}`).join(" ") +
    ` L ${xAt(n - 1).toFixed(2)} ${baseY} Z`;

  const gridSteps = [0, 0.5, 1];
  const midIndex = Math.floor((n - 1) / 2);
  const last = data[n - 1];

  function handlePointerMove(event: React.PointerEvent<SVGSVGElement>) {
    const svg = svgRef.current;
    if (!svg || n === 0) return;
    const rect = svg.getBoundingClientRect();
    const relativeX = ((event.clientX - rect.left) / rect.width) * WIDTH;
    const fraction = (relativeX - PAD.left) / innerWidth;
    const index = Math.round(fraction * (n - 1));
    setHoverIndex(Math.min(n - 1, Math.max(0, index)));
  }

  const hovered = hoverIndex !== null ? data[hoverIndex] : null;
  const hoveredX = hoverIndex !== null ? xAt(hoverIndex) : 0;
  const hoveredY = hovered ? yAt(hovered.revenue) : 0;
  const total = data.reduce((sum, d) => sum + d.revenue, 0);

  return (
    <div className="relative mt-4">
      <svg
        ref={svgRef}
        viewBox={`0 0 ${WIDTH} ${HEIGHT}`}
        preserveAspectRatio="none"
        className="h-56 w-full"
        onPointerMove={handlePointerMove}
        onPointerLeave={() => setHoverIndex(null)}
        role="img"
        aria-label={`Receita diária de ${formatDayLabel(data[0].date)} a ${formatDayLabel(
          last.date
        )}, total ${fullCurrency.format(total)}`}
      >
        {gridSteps.map((step) => {
          const y = PAD.top + innerHeight - step * innerHeight;
          return (
            <g key={step}>
              <line x1={PAD.left} x2={WIDTH - PAD.right} y1={y} y2={y} stroke={GRIDLINE_COLOR} strokeWidth={1} />
              <text x={PAD.left - 8} y={y} textAnchor="end" dominantBaseline="middle" fontSize={10} fill={MUTED_TEXT}>
                {axisCurrency.format(step * max)}
              </text>
            </g>
          );
        })}

        <text x={xAt(0)} y={HEIGHT - 8} textAnchor="start" fontSize={10} fill={MUTED_TEXT}>
          {formatDayLabel(data[0].date)}
        </text>
        {n > 2 && (
          <text x={xAt(midIndex)} y={HEIGHT - 8} textAnchor="middle" fontSize={10} fill={MUTED_TEXT}>
            {formatDayLabel(data[midIndex].date)}
          </text>
        )}
        <text x={xAt(n - 1)} y={HEIGHT - 8} textAnchor="end" fontSize={10} fill={MUTED_TEXT}>
          {formatDayLabel(last.date)}
        </text>

        <path d={areaPath} fill={SERIES_COLOR} fillOpacity={0.1} stroke="none" />
        <path d={linePath} fill="none" stroke={SERIES_COLOR} strokeWidth={2} strokeLinejoin="round" strokeLinecap="round" />
        <circle cx={xAt(n - 1)} cy={yAt(last.revenue)} r={5} fill={SERIES_COLOR} stroke={SURFACE_COLOR} strokeWidth={2} />

        {hovered && (
          <g>
            <line
              x1={hoveredX}
              x2={hoveredX}
              y1={PAD.top}
              y2={PAD.top + innerHeight}
              stroke={CROSSHAIR_COLOR}
              strokeWidth={1}
            />
            <circle cx={hoveredX} cy={hoveredY} r={5} fill={SERIES_COLOR} stroke={SURFACE_COLOR} strokeWidth={2} />
          </g>
        )}
      </svg>

      {hovered && (
        <div
          className="pointer-events-none absolute top-0 -translate-x-1/2 whitespace-nowrap rounded-md border border-gray-200 bg-white px-2.5 py-1.5 text-xs shadow-sm"
          style={{ left: `${(hoveredX / WIDTH) * 100}%` }}
        >
          <p className="text-gray-500">{formatDayLabel(hovered.date)}</p>
          <p className="font-semibold text-gray-900">{fullCurrency.format(hovered.revenue)}</p>
        </div>
      )}
    </div>
  );
}
