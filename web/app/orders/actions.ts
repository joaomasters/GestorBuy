"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { revalidatePath } from "next/cache";
import { apiFetch, ApiError } from "@/lib/api";
import { SESSION_COOKIE } from "@/lib/session";

export type ActionState = { error: string } | { synced: number } | null;

async function requireSession() {
  const cookieStore = await cookies();
  if (!cookieStore.has(SESSION_COOKIE)) {
    redirect("/login");
  }
}

export async function syncOrders(
  _prevState: ActionState,
  _formData: FormData
): Promise<ActionState> {
  await requireSession();

  try {
    const result = await apiFetch<{ synced: number }>("/orders/sync", {
      method: "POST",
    });
    revalidatePath("/orders");
    return { synced: result.synced };
  } catch (err) {
    if (err instanceof ApiError) return { error: err.message };
    return { error: "Não foi possível sincronizar os pedidos" };
  }
}
