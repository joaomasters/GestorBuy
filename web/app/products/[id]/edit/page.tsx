import { notFound } from "next/navigation";
import { apiFetch, ApiError } from "@/lib/api";
import { ProductForm, type ProductFormInitialData } from "../../product-form";
import { DeleteButton } from "../delete-button";
import { ChannelSection } from "../channel-section";

export default async function EditProductPage(
  props: PageProps<"/products/[id]/edit">
) {
  const { id } = await props.params;

  let product: ProductFormInitialData;
  try {
    product = await apiFetch<ProductFormInitialData>(`/products/${id}`);
  } catch (err) {
    if (err instanceof ApiError && err.status === 404) {
      notFound();
    }
    throw err;
  }

  return (
    <main className="mx-auto max-w-2xl px-4 py-10">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-semibold">Editar produto</h1>
        <DeleteButton productId={id} />
      </div>
      <ProductForm mode="edit" productId={id} initialData={product} />
      <ChannelSection productId={id} channel={product.channels?.mercadolivre} />
    </main>
  );
}
