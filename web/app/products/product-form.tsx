"use client";

import { useActionState, useState } from "react";
import { createProduct, updateProduct, type ActionState } from "./actions";

type VariationRow = {
  sku: string;
  cor: string;
  tamanho: string;
  stock: string;
  cost: string;
  price: string;
};

const emptyRow: VariationRow = {
  sku: "",
  cor: "",
  tamanho: "",
  stock: "0",
  cost: "0",
  price: "0",
};

export type ProductFormInitialData = {
  sku_master: string;
  title: string;
  brand?: string;
  category_normalized?: string;
  variations: {
    variation_sku: string;
    attributes?: Record<string, string>;
    stock_total: number;
    cost_price: number;
    price: number;
  }[];
};

type Props =
  | { mode: "create" }
  | { mode: "edit"; productId: string; initialData: ProductFormInitialData };

const initialState: ActionState = null;

export function ProductForm(props: Props) {
  const action =
    props.mode === "edit"
      ? updateProduct.bind(null, props.productId)
      : createProduct;
  const [state, formAction, pending] = useActionState(action, initialState);

  const initialData = props.mode === "edit" ? props.initialData : undefined;
  const [rows, setRows] = useState<VariationRow[]>(
    initialData && initialData.variations.length > 0
      ? initialData.variations.map((v) => ({
          sku: v.variation_sku,
          cor: v.attributes?.cor ?? "",
          tamanho: v.attributes?.tamanho ?? "",
          stock: String(v.stock_total),
          cost: String(v.cost_price),
          price: String(v.price),
        }))
      : [emptyRow]
  );

  function updateRow(index: number, patch: Partial<VariationRow>) {
    setRows((prev) => prev.map((r, i) => (i === index ? { ...r, ...patch } : r)));
  }

  function addRow() {
    setRows((prev) => [...prev, emptyRow]);
  }

  function removeRow(index: number) {
    setRows((prev) => prev.filter((_, i) => i !== index));
  }

  const inputClass =
    "rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-gray-500 focus:outline-none";
  const rowInputClass =
    "w-28 rounded-md border border-gray-300 px-2 py-1 text-sm focus:border-gray-500 focus:outline-none";

  return (
    <form action={formAction} className="mt-6 flex flex-col gap-6">
      {props.mode === "create" && (
        <label className="flex flex-col gap-1 text-sm">
          SKU mestre
          <input name="sku_master" required className={inputClass} />
        </label>
      )}

      <label className="flex flex-col gap-1 text-sm">
        Título
        <input
          name="title"
          required
          defaultValue={initialData?.title}
          className={inputClass}
        />
      </label>

      <div className="grid grid-cols-2 gap-4">
        <label className="flex flex-col gap-1 text-sm">
          Marca
          <input
            name="brand"
            defaultValue={initialData?.brand}
            className={inputClass}
          />
        </label>
        <label className="flex flex-col gap-1 text-sm">
          Categoria
          <input
            name="category_normalized"
            placeholder="ex.: vestuario.camisas.polo"
            defaultValue={initialData?.category_normalized}
            className={inputClass}
          />
        </label>
      </div>

      <div>
        <div className="flex items-center justify-between">
          <h2 className="text-sm font-medium text-gray-700">Variações</h2>
          <button
            type="button"
            onClick={addRow}
            className="text-sm text-gray-900 underline"
          >
            + adicionar variação
          </button>
        </div>

        <div className="mt-3 flex flex-col gap-3">
          {rows.map((row, i) => (
            <div
              key={i}
              className="flex flex-wrap items-center gap-2 rounded-md border border-gray-200 p-3"
            >
              <input
                placeholder="SKU"
                value={row.sku}
                onChange={(e) => updateRow(i, { sku: e.target.value })}
                className={rowInputClass}
              />
              <input
                placeholder="Cor"
                value={row.cor}
                onChange={(e) => updateRow(i, { cor: e.target.value })}
                className={rowInputClass}
              />
              <input
                placeholder="Tamanho"
                value={row.tamanho}
                onChange={(e) => updateRow(i, { tamanho: e.target.value })}
                className={rowInputClass}
              />
              <input
                placeholder="Estoque"
                type="number"
                value={row.stock}
                onChange={(e) => updateRow(i, { stock: e.target.value })}
                className={rowInputClass}
              />
              <input
                placeholder="Preço"
                type="number"
                step="0.01"
                value={row.price}
                onChange={(e) => updateRow(i, { price: e.target.value })}
                className={rowInputClass}
              />
              <input
                placeholder="Custo"
                type="number"
                step="0.01"
                value={row.cost}
                onChange={(e) => updateRow(i, { cost: e.target.value })}
                className={rowInputClass}
              />
              {rows.length > 1 && (
                <button
                  type="button"
                  onClick={() => removeRow(i)}
                  className="text-xs text-red-600"
                >
                  remover
                </button>
              )}
            </div>
          ))}
        </div>
      </div>

      <input type="hidden" name="variations_json" value={JSON.stringify(rows)} />

      {state?.error && <p className="text-sm text-red-600">{state.error}</p>}

      <button
        type="submit"
        disabled={pending}
        className="self-start rounded-md bg-gray-900 px-4 py-2 text-white disabled:opacity-50"
      >
        {pending ? "Salvando..." : "Salvar"}
      </button>
    </form>
  );
}
