import { redirect } from "next/navigation";
import { apiFetch, ApiError } from "@/lib/api";
import { RevenueChart } from "./revenue-chart";
import { RangeSelector } from "./range-selector";

type Summary = {
  from: string;
  to: string;
  revenue: number;
  gross_profit: number;
  margin_percent: number;
  unmatched_revenue: number;
  daily: { date: string; revenue: number }[];
};

const currency = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });

const RANGE_OPTIONS = [7, 30, 90] as const;
type RangeOption = (typeof RANGE_OPTIONS)[number];

function parseRange(raw: string | undefined): RangeOption {
  const parsed = Number(raw);
  return (RANGE_OPTIONS as readonly number[]).includes(parsed) ? (parsed as RangeOption) : 30;
}

function toDateParam(d: Date): string {
  return d.toISOString().slice(0, 10);
}

// Preenche os dias sem pedido com receita 0 — sem isso o gráfico ligaria
// dias distantes como se fossem consecutivos (a API só devolve dias que
// tiveram pedido, ver internal/dashboard/service.go).
function fillDailyGaps(
  daily: { date: string; revenue: number }[],
  from: Date,
  to: Date
): { date: string; revenue: number }[] {
  const byDate = new Map(daily.map((d) => [d.date, d.revenue]));
  const result: { date: string; revenue: number }[] = [];
  const cursor = new Date(Date.UTC(from.getUTCFullYear(), from.getUTCMonth(), from.getUTCDate()));
  const end = new Date(Date.UTC(to.getUTCFullYear(), to.getUTCMonth(), to.getUTCDate()));
  while (cursor <= end) {
    const key = toDateParam(cursor);
    result.push({ date: key, revenue: byDate.get(key) ?? 0 });
    cursor.setUTCDate(cursor.getUTCDate() + 1);
  }
  return result;
}

export default async function DashboardPage(props: PageProps<"/dashboard">) {
  const searchParams = await props.searchParams;
  const range = parseRange(
    typeof searchParams?.range === "string" ? searchParams.range : undefined
  );

  const to = new Date();
  const from = new Date(to);
  from.setUTCDate(from.getUTCDate() - (range - 1));

  let data: Summary;
  try {
    data = await apiFetch<Summary>(
      `/dashboard/summary?from=${toDateParam(from)}&to=${toDateParam(to)}`
    );
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/login");
    }
    throw err;
  }

  const daily = fillDailyGaps(data.daily, from, to);
  const hasRevenue = data.revenue > 0;
  const marginLabel = hasRevenue ? `${data.margin_percent.toFixed(1)}%` : "—";

  return (
    <main className="mx-auto max-w-4xl px-4 py-10">
      <div className="flex flex-wrap items-center justify-between gap-4">
        <div>
          <h1 className="text-2xl font-semibold">Dashboard</h1>
          <p className="mt-1 text-sm text-gray-500">
            Faturamento e lucro dos pedidos sincronizados em{" "}
            <a href="/orders" className="underline">
              Pedidos
            </a>
            .
          </p>
        </div>
        <RangeSelector current={range} options={RANGE_OPTIONS} />
      </div>

      <div className="mt-8 grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatTile label="Faturamento" value={currency.format(data.revenue)} />
        <StatTile label="Lucro bruto" value={currency.format(data.gross_profit)} />
        <StatTile label="Margem" value={marginLabel} />
      </div>

      {data.unmatched_revenue > 0 && (
        <p className="mt-4 rounded-md border border-amber-200 bg-amber-50 px-3 py-2 text-sm text-amber-800">
          {currency.format(data.unmatched_revenue)} em vendas não entraram no lucro bruto porque
          o anúncio ainda não está vinculado a um produto com custo cadastrado.{" "}
          <a href="/products" className="underline">
            Vincular produtos
          </a>
          .
        </p>
      )}

      <div className="mt-8 rounded-lg border border-gray-200 bg-white p-4">
        <h2 className="text-sm font-medium text-gray-700">Receita diária</h2>
        {daily.every((d) => d.revenue === 0) ? (
          <p className="mt-4 py-12 text-center text-sm text-gray-500">
            Nenhum pedido sincronizado nesse período.
          </p>
        ) : (
          <RevenueChart data={daily} />
        )}
      </div>
    </main>
  );
}

function StatTile({ label, value }: { label: string; value: string }) {
  return (
    <div className="rounded-lg border border-gray-200 bg-white p-4">
      <p className="text-sm text-gray-500">{label}</p>
      <p className="mt-1 text-2xl font-semibold">{value}</p>
    </div>
  );
}
