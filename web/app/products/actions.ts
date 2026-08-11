"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { revalidatePath } from "next/cache";
import { apiFetch, ApiError } from "@/lib/api";
import { SESSION_COOKIE } from "@/lib/session";

export type ActionState = { error: string } | null;

// Defesa em profundidade: o proxy.ts já bloqueia /products/* sem cookie,
// mas Server Actions são alcançáveis por POST direto e não dependem só do
// proxy (ver aviso da doc do Next.js sobre matcher de proxy não cobrir
// Server Functions se o path mudar) — cada action confirma a sessão de
// novo. A autorização de verdade continua sendo o backend Go validando o
// JWT em cada chamada.
async function requireSession() {
  const cookieStore = await cookies();
  if (!cookieStore.has(SESSION_COOKIE)) {
    redirect("/login");
  }
}

type RawVariationRow = {
  sku: string;
  cor: string;
  tamanho: string;
  stock: string;
  cost: string;
  price: string;
};

function parseVariations(formData: FormData) {
  const raw = formData.get("variations_json");
  if (typeof raw !== "string" || raw === "") return [];

  let rows: RawVariationRow[];
  try {
    rows = JSON.parse(raw);
  } catch {
    return [];
  }

  return rows
    .filter((row) => row.sku.trim() !== "")
    .map((row) => {
      const attributes: Record<string, string> = {};
      if (row.cor.trim()) attributes.cor = row.cor.trim();
      if (row.tamanho.trim()) attributes.tamanho = row.tamanho.trim();

      return {
        variation_sku: row.sku.trim(),
        attributes: Object.keys(attributes).length ? attributes : undefined,
        stock_total: Number(row.stock) || 0,
        cost_price: Number(row.cost) || 0,
        price: Number(row.price) || 0,
      };
    });
}

function errorMessage(err: unknown, fallback: string): ActionState {
  if (err instanceof ApiError) return { error: err.message };
  return { error: fallback };
}

export async function createProduct(
  _prevState: ActionState,
  formData: FormData
): Promise<ActionState> {
  await requireSession();

  const payload = {
    sku_master: String(formData.get("sku_master") ?? ""),
    title: String(formData.get("title") ?? ""),
    brand: String(formData.get("brand") ?? ""),
    category_normalized: String(formData.get("category_normalized") ?? ""),
    variations: parseVariations(formData),
  };

  try {
    await apiFetch("/products", {
      method: "POST",
      body: JSON.stringify(payload),
    });
  } catch (err) {
    return errorMessage(err, "Não foi possível criar o produto");
  }

  revalidatePath("/products");
  redirect("/products");
}

export async function updateProduct(
  id: string,
  _prevState: ActionState,
  formData: FormData
): Promise<ActionState> {
  await requireSession();

  const payload = {
    title: String(formData.get("title") ?? ""),
    brand: String(formData.get("brand") ?? ""),
    category_normalized: String(formData.get("category_normalized") ?? ""),
    variations: parseVariations(formData),
  };

  try {
    await apiFetch(`/products/${id}`, {
      method: "PATCH",
      body: JSON.stringify(payload),
    });
  } catch (err) {
    return errorMessage(err, "Não foi possível atualizar o produto");
  }

  revalidatePath("/products");
  redirect("/products");
}

export async function deleteProduct(id: string, _formData: FormData) {
  await requireSession();

  try {
    await apiFetch(`/products/${id}`, { method: "DELETE" });
  } catch (err) {
    if (err instanceof ApiError) {
      throw new Error(err.message);
    }
    throw err;
  }

  revalidatePath("/products");
  redirect("/products");
}
