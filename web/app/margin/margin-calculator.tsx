"use client";

import { useActionState, useState } from "react";
import { calculateMargin, type ActionState } from "./actions";

type ProductOption = {
  id: string;
  title: string;
  sku_master: string;
  costPrice: number;
  price: number;
};

const initialState: ActionState = null;
const currency = new Intl.NumberFormat("pt-BR", { style: "currency", currency: "BRL" });

export function MarginCalculator({ products }: { products: ProductOption[] }) {
  const [state, formAction, pending] = useActionState(calculateMargin, initialState);
  const [costPrice, setCostPrice] = useState("");
  const [salePrice, setSalePrice] = useState("");

  function handleProductSelect(id: string) {
    const product = products.find((p) => p.id === id);
    if (product) {
      setCostPrice(String(product.costPrice));
      setSalePrice(String(product.price));
    }
  }

  const inputClass =
    "rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-gray-500 focus:outline-none";

  return (
    <form action={formAction} className="mt-6 flex flex-col gap-6">
      {products.length > 0 && (
        <label className="flex flex-col gap-1 text-sm">
          Preencher a partir de um produto do catálogo (opcional)
          <select
            onChange={(e) => handleProductSelect(e.target.value)}
            defaultValue=""
            className={inputClass}
          >
            <option value="">Selecionar...</option>
            {products.map((p) => (
              <option key={p.id} value={p.id}>
                {p.title} ({p.sku_master})
              </option>
            ))}
          </select>
        </label>
      )}

      <div className="grid grid-cols-2 gap-4">
        <label className="flex flex-col gap-1 text-sm">
          Custo do produto (R$)
          <input
            name="cost_price"
            type="number"
            step="0.01"
            required
            value={costPrice}
            onChange={(e) => setCostPrice(e.target.value)}
            className={inputClass}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          Preço de venda (R$)
          <input
            name="sale_price"
            type="number"
            step="0.01"
            required
            value={salePrice}
            onChange={(e) => setSalePrice(e.target.value)}
            className={inputClass}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          Taxa do marketplace (%)
          <input
            name="marketplace_fee_percent"
            type="number"
            step="0.01"
            defaultValue="12"
            className={inputClass}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          Frete (R$)
          <input name="shipping_cost" type="number" step="0.01" defaultValue="0" className={inputClass} />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          Imposto (%)
          <input name="tax_percent" type="number" step="0.01" defaultValue="0" className={inputClass} />
        </label>
      </div>

      <button
        type="submit"
        disabled={pending}
        className="self-start rounded-md bg-gray-900 px-4 py-2 text-sm text-white disabled:opacity-50"
      >
        {pending ? "Calculando..." : "Calcular"}
      </button>

      {state && "error" in state && (
        <p className="text-sm text-red-600">{state.error}</p>
      )}

      {state && "result" in state && (
        <div className="rounded-md border border-gray-200 p-4">
          <dl className="grid grid-cols-2 gap-3 text-sm">
            <dt className="text-gray-500">Taxa do marketplace</dt>
            <dd className="text-right">{currency.format(state.result.marketplace_fee_amount)}</dd>
            <dt className="text-gray-500">Imposto</dt>
            <dd className="text-right">{currency.format(state.result.tax_amount)}</dd>
            <dt className="font-medium text-gray-700">Lucro líquido</dt>
            <dd
              className={`text-right font-medium ${state.result.net_profit >= 0 ? "text-green-700" : "text-red-600"}`}
            >
              {currency.format(state.result.net_profit)}
            </dd>
            <dt className="font-medium text-gray-700">Margem</dt>
            <dd
              className={`text-right font-medium ${state.result.margin_percent >= 0 ? "text-green-700" : "text-red-600"}`}
            >
              {state.result.margin_percent.toFixed(1)}%
            </dd>
            <dt className="text-gray-500">Preço de equilíbrio</dt>
            <dd className="text-right">
              {state.result.break_even_price != null
                ? currency.format(state.result.break_even_price)
                : "Impossível (taxa + imposto ≥ 100%)"}
            </dd>
          </dl>
        </div>
      )}
    </form>
  );
}
