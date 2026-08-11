import { redirect } from "next/navigation";
import { apiFetch, ApiError } from "@/lib/api";
import { SyncButton } from "./sync-button";

type OrderItem = { title: string; quantity: number; unit_price: number };

type Order = {
  id: string;
  marketplace: string;
  external_order_id: string;
  status: string;
  total_amount: number;
  buyer_nickname?: string;
  items: OrderItem[];
  date_created: string;
};

type ListResponse = { orders: Order[] };

const currency = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });

const statusLabels: Record<string, string> = {
  paid: "Pago",
  confirmed: "Confirmado",
  cancelled: "Cancelado",
  invalid: "Inválido",
};

export default async function OrdersPage() {
  let data: ListResponse;
  try {
    data = await apiFetch<ListResponse>("/orders");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/login");
    }
    throw err;
  }

  return (
    <main className="mx-auto max-w-4xl px-4 py-10">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-semibold">Pedidos</h1>
          <p className="mt-1 text-sm text-gray-500">
            Pedidos da conta do Mercado Livre conectada em{" "}
            <a href="/integrations" className="underline">
              Integrações
            </a>
            .
          </p>
        </div>
        <SyncButton />
      </div>

      {data.orders.length === 0 ? (
        <p className="mt-8 text-gray-500">
          Nenhum pedido sincronizado ainda — clique em &quot;Sincronizar
          agora&quot;.
        </p>
      ) : (
        <table className="mt-8 w-full text-left text-sm">
          <thead>
            <tr className="border-b border-gray-200 text-gray-500">
              <th className="py-2 font-medium">Pedido</th>
              <th className="py-2 font-medium">Comprador</th>
              <th className="py-2 font-medium">Itens</th>
              <th className="py-2 font-medium">Status</th>
              <th className="py-2 font-medium">Total</th>
            </tr>
          </thead>
          <tbody>
            {data.orders.map((order) => (
              <tr key={order.id} className="border-b border-gray-100">
                <td className="py-3 font-mono text-xs">{order.external_order_id}</td>
                <td className="py-3">{order.buyer_nickname ?? "—"}</td>
                <td className="py-3">
                  {order.items.map((it) => `${it.quantity}x ${it.title}`).join(", ")}
                </td>
                <td className="py-3">{statusLabels[order.status] ?? order.status}</td>
                <td className="py-3">{currency.format(order.total_amount)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </main>
  );
}
