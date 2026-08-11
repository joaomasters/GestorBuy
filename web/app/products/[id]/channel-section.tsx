"use client";

import { useActionState } from "react";
import { linkChannel, syncChannel, type ActionState } from "../actions";
import type { Channel } from "../product-form";

const MARKETPLACE = "mercadolivre";
const initialState: ActionState = null;

const statusLabels: Record<string, string> = {
  linked: "Vinculado, ainda não sincronizado",
  synced: "Sincronizado",
  error: "Erro na última sincronização",
};

export function ChannelSection({
  productId,
  channel,
}: {
  productId: string;
  channel?: Channel;
}) {
  const [linkState, linkAction, linkPending] = useActionState(
    linkChannel.bind(null, productId, MARKETPLACE),
    initialState
  );
  const [syncState, syncAction, syncPending] = useActionState(
    syncChannel.bind(null, productId, MARKETPLACE),
    initialState
  );

  const isLinked = Boolean(channel?.item_id);

  return (
    <div className="mt-10 rounded-md border border-gray-200 p-4">
      <h2 className="text-sm font-medium text-gray-700">Mercado Livre</h2>

      {isLinked ? (
        <div className="mt-3 flex flex-col gap-2 text-sm">
          <p>
            Anúncio vinculado: <span className="font-mono">{channel!.item_id}</span>
          </p>
          <p className="text-gray-500">
            {statusLabels[channel!.sync_status ?? ""] ?? "Status desconhecido"}
            {channel!.last_synced_at &&
              ` — última sincronização em ${new Date(channel!.last_synced_at).toLocaleString("pt-BR")}`}
          </p>
          {channel!.last_error && (
            <p className="text-red-600">Erro: {channel!.last_error}</p>
          )}

          <form action={syncAction} className="mt-1">
            <button
              type="submit"
              disabled={syncPending}
              className="rounded-md bg-gray-900 px-3 py-1.5 text-sm text-white disabled:opacity-50"
            >
              {syncPending ? "Sincronizando..." : "Sincronizar agora"}
            </button>
          </form>
          {syncState?.error && (
            <p className="text-sm text-red-600">{syncState.error}</p>
          )}
        </div>
      ) : (
        <form action={linkAction} className="mt-3 flex items-end gap-2">
          <label className="flex flex-col gap-1 text-sm">
            ID do anúncio (ex.: MLB1234567890)
            <input
              name="item_id"
              required
              className="w-56 rounded-md border border-gray-300 px-3 py-2 text-sm"
            />
          </label>
          <button
            type="submit"
            disabled={linkPending}
            className="rounded-md bg-gray-900 px-3 py-2 text-sm text-white disabled:opacity-50"
          >
            {linkPending ? "Vinculando..." : "Vincular"}
          </button>
        </form>
      )}
      {linkState?.error && (
        <p className="mt-2 text-sm text-red-600">{linkState.error}</p>
      )}
      <p className="mt-3 text-xs text-gray-400">
        Requer a integração com o Mercado Livre conectada em{" "}
        <a href="/integrations" className="underline">
          Integrações
        </a>
        .
      </p>
    </div>
  );
}
