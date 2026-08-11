"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { revalidatePath } from "next/cache";
import { apiFetch, ApiError } from "@/lib/api";
import { SESSION_COOKIE } from "@/lib/session";

export type ActionState = { error: string } | null;

async function requireSession() {
  const cookieStore = await cookies();
  if (!cookieStore.has(SESSION_COOKIE)) {
    redirect("/login");
  }
}

export async function trackItem(
  _prevState: ActionState,
  formData: FormData
): Promise<ActionState> {
  await requireSession();

  const urlOrID = String(formData.get("url_or_id") ?? "").trim();
  if (!urlOrID) {
    return { error: "Cola o link ou o ID do anúncio" };
  }

  try {
    await apiFetch("/mining/items", {
      method: "POST",
      body: JSON.stringify({ url_or_id: urlOrID }),
    });
  } catch (err) {
    if (err instanceof ApiError) return { error: err.message };
    return { error: "Não foi possível rastrear esse anúncio" };
  }

  revalidatePath("/mining");
  return null;
}

export async function refreshItem(id: string, _formData: FormData) {
  await requireSession();

  try {
    await apiFetch(`/mining/items/${id}/refresh`, { method: "POST" });
  } catch (err) {
    if (err instanceof ApiError) {
      throw new Error(err.message);
    }
    throw err;
  }

  revalidatePath("/mining");
}
