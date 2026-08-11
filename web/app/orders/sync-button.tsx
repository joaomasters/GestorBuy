"use client";

import { useActionState } from "react";
import { syncOrders, type ActionState } from "./actions";

const initialState: ActionState = null;

export function SyncButton() {
  const [state, formAction, pending] = useActionState(syncOrders, initialState);

  return (
    <div className="flex flex-col items-end gap-1">
      <form action={formAction}>
        <button
          type="submit"
          disabled={pending}
          className="rounded-md bg-gray-900 px-4 py-2 text-sm text-white disabled:opacity-50"
        >
          {pending ? "Sincronizando..." : "Sincronizar agora"}
        </button>
      </form>
      {state && "error" in state && (
        <p className="text-sm text-red-600">{state.error}</p>
      )}
      {state && "synced" in state && (
        <p className="text-sm text-green-700">{state.synced} pedido(s) sincronizado(s)</p>
      )}
    </div>
  );
}
