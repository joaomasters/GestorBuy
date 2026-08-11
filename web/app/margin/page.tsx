import { redirect } from "next/navigation";
import { apiFetch, ApiError } from "@/lib/api";
import { MarginCalculator } from "./margin-calculator";

type Product = {
  id: string;
  sku_master: string;
  title: string;
  variations: { cost_price: number; price: number }[];
};

type ListResponse = { items: Product[] };

export default async function MarginPage() {
  let data: ListResponse;
  try {
    data = await apiFetch<ListResponse>("/products?limit=100");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/login");
    }
    throw err;
  }

  const products = data.items
    .filter((p) => p.variations.length > 0)
    .map((p) => ({
      id: p.id,
      title: p.title,
      sku_master: p.sku_master,
      costPrice: p.variations[0].cost_price,
      price: p.variations[0].price,
    }));

  return (
    <main className="mx-auto max-w-2xl px-4 py-10">
      <h1 className="text-2xl font-semibold">Calculadora de margem</h1>
      <p className="mt-1 text-sm text-gray-500">
        Confere se o preço cobre taxa do marketplace, frete e imposto antes
        de publicar ou ajustar um anúncio.
      </p>
      <MarginCalculator products={products} />
    </main>
  );
}
