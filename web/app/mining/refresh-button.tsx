"use client";

import { refreshItem } from "./actions";

export function RefreshButton({ itemId }: { itemId: string }) {
  return (
    <form action={refreshItem.bind(null, itemId)}>
      <button type="submit" className="text-sm text-gray-900 underline">
        atualizar agora
      </button>
    </form>
  );
}
