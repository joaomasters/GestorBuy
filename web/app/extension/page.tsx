export default function ExtensionPage() {
  return (
    <main className="mx-auto max-w-2xl px-4 py-10">
      <h1 className="text-2xl font-semibold">Extensão de navegador</h1>
      <p className="mt-2 text-sm text-gray-500">
        A mineração pela API (formulário em{" "}
        <a href="/mining" className="underline">
          Mineração
        </a>
        ) só funciona pros seus próprios anúncios — o Mercado Livre não
        libera leitura de item de terceiro pra apps autenticados. Pra minerar
        anúncio de concorrente, a extensão lê os dados direto da página
        renderizada, na sua própria sessão de navegação — pro Mercado Livre,
        isso é indistinguível de você navegando normalmente.
      </p>

      <div className="mt-6 rounded-md border border-gray-200 bg-white p-4 text-sm">
        <h2 className="font-medium">Instalar (Chrome/Edge, modo desenvolvedor)</h2>
        <p className="mt-2 text-gray-600">
          A extensão ainda não está na Chrome Web Store (isso exige revisão
          do Google) — por enquanto é instalada manualmente, mas sem precisar
          instalar nada além do navegador.
        </p>

        <a
          href="/gestorbuy-extension.zip"
          download
          className="mt-4 inline-block rounded-md bg-gray-900 px-4 py-2 text-sm text-white"
        >
          Baixar extensão (.zip)
        </a>

        <ol className="mt-4 list-decimal space-y-1 pl-5 text-gray-600">
          <li>Baixe o .zip acima e extraia numa pasta (ex.: Downloads\gestorbuy-extension)</li>
          <li>
            Abra <code className="rounded bg-gray-100 px-1 py-0.5">chrome://extensions</code> (ou{" "}
            <code className="rounded bg-gray-100 px-1 py-0.5">edge://extensions</code>) e ative o
            &quot;Modo do desenvolvedor&quot; (canto superior direito)
          </li>
          <li>
            Clique em &quot;Carregar sem compactação&quot; e selecione a pasta que você extraiu
          </li>
          <li>O ícone do GestorBuy aparece na barra de extensões — fixe-o pra achar fácil depois</li>
          <li>Clique no ícone, faça login com seu e-mail/senha do GestorBuy</li>
          <li>
            Navegue até qualquer anúncio no <code className="rounded bg-gray-100 px-1 py-0.5">mercadolivre.com.br</code>
            {" "}(seu ou de concorrente) e clique no ícone de novo
          </li>
          <li>Confira a prévia (título, preço, vendidos, vendedor) e clique em &quot;Salvar no GestorBuy&quot;</li>
        </ol>

        <p className="mt-4 text-xs text-gray-400">
          O item salvo aparece em{" "}
          <a href="/mining" className="underline">
            Mineração
          </a>
          . Pra atualizar depois de uma mudança de preço, é só abrir o anúncio de
          novo e clicar em &quot;Salvar&quot; outra vez.
        </p>
      </div>

      <div className="mt-4 rounded-md border border-gray-200 bg-white p-4 text-xs text-gray-500">
        <p>
          Prefere buildar do código-fonte (pra desenvolver/contribuir)? Instruções em{" "}
          <code className="rounded bg-gray-100 px-1 py-0.5">extension/README.md</code> no repositório.
        </p>
      </div>
    </main>
  );
}
