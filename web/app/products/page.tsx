import Link from "next/link";
import { redirect } from "next/navigation";
import { apiFetch, ApiError } from "@/lib/api";

type Product = {
  id: string;
  sku_master: string;
  title: string;
  brand?: string;
  variations: { variation_sku: string; stock_total: number; price: number }[];
};

type ListResponse = {
  items: Product[];
  total: number;
  limit: number;
  offset: number;
};

export default async function ProductsPage() {
  let data: ListResponse;
  try {
    data = await apiFetch<ListResponse>("/products?limit=50");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/login");
    }
    throw err;
  }

  return (
    <main className="mx-auto max-w-4xl px-4 py-10">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Produtos</h1>
        <Link
          href="/products/new"
          className="rounded-md bg-gray-900 px-4 py-2 text-sm text-white"
        >
          Novo produto
        </Link>
      </div>

      {data.items.length === 0 ? (
        <p className="mt-8 text-gray-500">
          Nenhum produto cadastrado ainda.
        </p>
      ) : (
        <table className="mt-8 w-full text-left text-sm">
          <thead>
            <tr className="border-b border-gray-200 text-gray-500">
              <th className="py-2 font-medium">SKU</th>
              <th className="py-2 font-medium">Título</th>
              <th className="py-2 font-medium">Estoque total</th>
              <th className="py-2" />
            </tr>
          </thead>
          <tbody>
            {data.items.map((product) => {
              const totalStock = product.variations.reduce(
                (sum, v) => sum + v.stock_total,
                0
              );
              return (
                <tr key={product.id} className="border-b border-gray-100">
                  <td className="py-3 font-mono text-xs">
                    {product.sku_master}
                  </td>
                  <td className="py-3">{product.title}</td>
                  <td className="py-3">{totalStock}</td>
                  <td className="py-3 text-right">
                    <Link
                      href={`/products/${product.id}/edit`}
                      className="text-gray-900 underline"
                    >
                      Editar
                    </Link>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
    </main>
  );
}
