"use client";

import { useActionState } from "react";
import { trackItem, type ActionState } from "./actions";

const initialState: ActionState = null;

export function TrackForm() {
  const [state, formAction, pending] = useActionState(trackItem, initialState);

  return (
    <form action={formAction} className="flex items-end gap-2">
      <label className="flex flex-col gap-1 text-sm">
        Link ou ID de um anúncio seu no Mercado Livre
        <input
          name="url_or_id"
          required
          placeholder="https://produto.mercadolivre.com.br/MLB-... ou MLB1234567890"
          className="w-96 rounded-md border border-gray-300 px-3 py-2 text-sm"
        />
      </label>
      <button
        type="submit"
        disabled={pending}
        className="rounded-md bg-gray-900 px-4 py-2 text-sm text-white disabled:opacity-50"
      >
        {pending ? "Buscando..." : "Rastrear"}
      </button>
      {state?.error && (
        <p className="mb-2 text-sm text-red-600">{state.error}</p>
      )}
    </form>
  );
}
