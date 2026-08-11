"use client";

import { deleteProduct } from "../actions";

export function DeleteButton({ productId }: { productId: string }) {
  return (
    <form action={deleteProduct.bind(null, productId)}>
      <button type="submit" className="text-sm text-red-600 underline">
        Excluir
      </button>
    </form>
  );
}
