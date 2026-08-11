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
        <h2 className="font-medium">Ainda não está disponível pra instalar direto</h2>
        <p className="mt-2 text-gray-600">
          A extensão ainda não foi publicada na Chrome Web Store — precisa ser
          buildada e carregada em modo desenvolvedor a partir do código-fonte,
          na pasta <code className="rounded bg-gray-100 px-1 py-0.5">extension/</code>{" "}
          do repositório.
        </p>
        <ol className="mt-3 list-decimal space-y-1 pl-5 text-gray-600">
          <li>
            <code className="rounded bg-gray-100 px-1 py-0.5">cd extension && npm install && npm run build</code>
          </li>
          <li>
            Abra <code className="rounded bg-gray-100 px-1 py-0.5">chrome://extensions</code>, ative o
            &quot;Modo do desenvolvedor&quot;
          </li>
          <li>
            &quot;Carregar sem compactação&quot; → selecione a pasta{" "}
            <code className="rounded bg-gray-100 px-1 py-0.5">extension/</code>
          </li>
          <li>Faça login com sua conta GestorBuy no popup da extensão</li>
          <li>Navegue até qualquer anúncio no Mercado Livre e clique em &quot;Salvar no GestorBuy&quot;</li>
        </ol>
        <p className="mt-3 text-xs text-gray-400">
          Instruções completas em <code className="rounded bg-gray-100 px-1 py-0.5">extension/README.md</code> no
          repositório.
        </p>
      </div>
    </main>
  );
}
