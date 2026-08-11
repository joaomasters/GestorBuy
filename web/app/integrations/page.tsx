import { redirect } from "next/navigation";
import { apiFetch, ApiError } from "@/lib/api";
import { DisconnectButton } from "./disconnect-button";

type IntegrationStatus = {
  marketplace: string;
  connected: boolean;
  connected_at?: string;
};

type StatusResponse = { integrations: IntegrationStatus[] };

const marketplaceLabels: Record<string, string> = {
  mercado_livre: "Mercado Livre",
};

const errorMessages: Record<string, string> = {
  missing_code_or_state: "O Mercado Livre não retornou os dados esperados. Tenta conectar de novo.",
  connect_failed: "Não foi possível concluir a conexão com o Mercado Livre.",
};

export default async function IntegrationsPage(
  props: PageProps<"/integrations">
) {
  const searchParams = await props.searchParams;

  let data: StatusResponse;
  try {
    data = await apiFetch<StatusResponse>("/integrations");
  } catch (err) {
    if (err instanceof ApiError && err.status === 401) {
      redirect("/login");
    }
    throw err;
  }

  const connectedFlag = searchParams?.connected === "1";
  const errorFlag = typeof searchParams?.error === "string" ? searchParams.error : null;

  return (
    <main className="mx-auto max-w-4xl px-4 py-10">
      <h1 className="text-2xl font-semibold">Integrações</h1>
      <p className="mt-1 text-sm text-gray-500">
        Conecte suas contas de marketplace pra sincronizar preço e estoque
        dos produtos do catálogo.
      </p>

      {connectedFlag && (
        <p className="mt-4 rounded-md bg-green-50 px-4 py-2 text-sm text-green-700">
          Conectado com sucesso.
        </p>
      )}
      {errorFlag && (
        <p className="mt-4 rounded-md bg-red-50 px-4 py-2 text-sm text-red-700">
          {errorMessages[errorFlag] ?? "Não foi possível conectar."}
        </p>
      )}

      <div className="mt-8 flex flex-col gap-3">
        {data.integrations.map((integration) => (
          <div
            key={integration.marketplace}
            className="flex items-center justify-between rounded-md border border-gray-200 bg-white p-4"
          >
            <div>
              <p className="font-medium">
                {marketplaceLabels[integration.marketplace] ?? integration.marketplace}
              </p>
              <p className="text-sm text-gray-500">
                {integration.connected
                  ? `Conectado${integration.connected_at ? " em " + new Date(integration.connected_at).toLocaleDateString("pt-BR") : ""}`
                  : "Não conectado"}
              </p>
            </div>
            {integration.connected ? (
              <DisconnectButton marketplace={integration.marketplace} />
            ) : (
              <a
                href="/api/integrations/mercadolivre/connect"
                className="rounded-md bg-gray-900 px-4 py-2 text-sm text-white"
              >
                Conectar
              </a>
            )}
          </div>
        ))}
      </div>
    </main>
  );
}
