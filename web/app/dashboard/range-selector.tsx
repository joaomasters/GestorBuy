import Link from "next/link";

// Filtro de período — uma linha só, acima do conteúdo que ele escopa (KPIs
// e gráfico reagem juntos ao mesmo range). Só presets, sem date picker
// customizado (fora de escopo desta entrega).
export function RangeSelector({
  current,
  options,
}: {
  current: number;
  options: readonly number[];
}) {
  return (
    <div className="flex items-center gap-1 rounded-md border border-gray-200 bg-white p-1 text-sm">
      {options.map((days) => {
        const active = days === current;
        return (
          <Link
            key={days}
            href={`/dashboard?range=${days}`}
            className={
              active
                ? "rounded px-3 py-1 font-medium bg-gray-900 text-white"
                : "rounded px-3 py-1 text-gray-600 hover:bg-gray-100"
            }
          >
            {days} dias
          </Link>
        );
      })}
    </div>
  );
}
