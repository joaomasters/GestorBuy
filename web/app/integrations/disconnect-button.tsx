"use client";

import { disconnectIntegration } from "./actions";

export function DisconnectButton({ marketplace }: { marketplace: string }) {
  return (
    <form action={disconnectIntegration.bind(null, marketplace)}>
      <button type="submit" className="text-sm text-red-600 underline">
        Desconectar
      </button>
    </form>
  );
}
