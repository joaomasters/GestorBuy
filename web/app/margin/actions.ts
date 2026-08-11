"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { apiFetch, ApiError } from "@/lib/api";
import { SESSION_COOKIE } from "@/lib/session";

export type MarginResult = {
  marketplace_fee_amount: number;
  tax_amount: number;
  net_profit: number;
  margin_percent: number;
  break_even_price?: number;
};

export type ActionState = { result: MarginResult } | { error: string } | null;

async function requireSession() {
  const cookieStore = await cookies();
  if (!cookieStore.has(SESSION_COOKIE)) {
    redirect("/login");
  }
}

function num(formData: FormData, key: string): number {
  const raw = formData.get(key);
  const parsed = Number(raw);
  return Number.isFinite(parsed) ? parsed : 0;
}

export async function calculateMargin(
  _prevState: ActionState,
  formData: FormData
): Promise<ActionState> {
  await requireSession();

  const payload = {
    cost_price: num(formData, "cost_price"),
    sale_price: num(formData, "sale_price"),
    marketplace_fee_percent: num(formData, "marketplace_fee_percent"),
    shipping_cost: num(formData, "shipping_cost"),
    tax_percent: num(formData, "tax_percent"),
  };

  try {
    const result = await apiFetch<MarginResult>("/margin/calculate", {
      method: "POST",
      body: JSON.stringify(payload),
    });
    return { result };
  } catch (err) {
    if (err instanceof ApiError) return { error: err.message };
    return { error: "Não foi possível calcular a margem" };
  }
}
