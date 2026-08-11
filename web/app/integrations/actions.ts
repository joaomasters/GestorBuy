"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { revalidatePath } from "next/cache";
import { apiFetch, ApiError } from "@/lib/api";
import { SESSION_COOKIE } from "@/lib/session";

async function requireSession() {
  const cookieStore = await cookies();
  if (!cookieStore.has(SESSION_COOKIE)) {
    redirect("/login");
  }
}

// Só o Mercado Livre existe por enquanto, mas a assinatura já recebe
// `marketplace` — quando Shopee/Amazon/etc entrarem, o backend expõe
// DELETE /integrations/{marketplace} e essa action já funciona sem mudar.
export async function disconnectIntegration(marketplace: string, _formData: FormData) {
  await requireSession();

  try {
    await apiFetch(`/integrations/${marketplace}`, { method: "DELETE" });
  } catch (err) {
    if (err instanceof ApiError) {
      throw new Error(err.message);
    }
    throw err;
  }

  revalidatePath("/integrations");
}
