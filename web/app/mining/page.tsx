import { redirect } from "next/navigation";
import { apiFetch, ApiError } from "@/lib/api";
import { TrackForm } from "./track-form";
import { RefreshButton } from "./refresh-button";

type Snapshot = {
  ts: string;
  price: number;
  sold_quantity: number;
  available_quantity: number;
  estimated_revenue_total: number;
};

type MinedItem = {
  id: string;
  marketplace: string;
  external_item_id: string;
  title: string;
  permalink?: string;
  seller_nickname?: string;
  source?: string;
  latest_snapshot?: Snapshot;
};

type ListResponse = { items: MinedItem[] };

const currency = new Intl.NumberFormat("pt-BR", {
  style: "currency",
  currency: "BRL",
});

export default async function MiningPage() {
  let data: ListResponse;
  try {
    data = await apiFetch<ListResponse>("/mining/items");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/login");
    }
    throw err;
  }

  return (
    <main className="mx-auto max-w-4xl px-4 py-10">
      <h1 className="text-2xl font-semibold">Mineração</h1>
      <p className="mt-1 text-sm text-gray-500">
        Acompanhe preço, unidades vendidas e faturamento estimado ao longo do
        tempo.
      </p>
      <p className="mt-1 text-xs text-gray-400">
        Pelo formulário abaixo, só dá pra rastrear anúncios da própria conta
        conectada em{" "}
        <a href="/integrations" className="underline">
          Integrações
        </a>{" "}
        — a API do Mercado Livre não libera leitura de item de terceiro.
        Pra minerar anúncio de concorrente, use a{" "}
        <a href="/extension" className="underline">
          extensão de navegador
        </a>
        , que lê a página na sua própria sessão.
      </p>

      <div className="mt-6">
        <TrackForm />
      </div>

      {data.items.length === 0 ? (
        <p className="mt-8 text-gray-500">Nenhum anúncio rastreado ainda.</p>
      ) : (
        <table className="mt-8 w-full text-left text-sm">
          <thead>
            <tr className="border-b border-gray-200 text-gray-500">
              <th className="py-2 font-medium">Anúncio</th>
              <th className="py-2 font-medium">Origem</th>
              <th className="py-2 font-medium">Preço</th>
              <th className="py-2 font-medium">Vendidos</th>
              <th className="py-2 font-medium">Faturamento estimado</th>
              <th className="py-2" />
            </tr>
          </thead>
          <tbody>
            {data.items.map((item) => (
              <tr key={item.id} className="border-b border-gray-100">
                <td className="py-3">
                  {item.permalink ? (
                    <a
                      href={item.permalink}
                      target="_blank"
                      rel="noreferrer"
                      className="underline"
                    >
                      {item.title}
                    </a>
                  ) : (
                    item.title
                  )}
                  <div className="font-mono text-xs text-gray-400">
                    {item.external_item_id}
                    {item.seller_nickname ? ` · ${item.seller_nickname}` : ""}
                  </div>
                </td>
                <td className="py-3 text-xs text-gray-500">
                  {item.source === "extension" ? "Extensão" : "Conta própria"}
                </td>
                <td className="py-3">
                  {item.latest_snapshot ? currency.format(item.latest_snapshot.price) : "—"}
                </td>
                <td className="py-3">
                  {item.latest_snapshot?.sold_quantity ?? "—"}
                </td>
                <td className="py-3">
                  {item.latest_snapshot
                    ? currency.format(item.latest_snapshot.estimated_revenue_total)
                    : "—"}
                </td>
                <td className="py-3 text-right">
                  {item.source === "extension" ? (
                    <span className="text-xs text-gray-400">atualize pela extensão</span>
                  ) : (
                    <RefreshButton itemId={item.id} />
                  )}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </main>
  );
}
