import Link from "next/link";
import { LogoutButton } from "./logout-button";

export function AppHeader() {
  return (
    <header className="border-b border-gray-200 bg-white">
      <div className="mx-auto flex max-w-4xl items-center justify-between px-4 py-4">
        <div className="flex items-center gap-6">
          <Link href="/products" className="font-semibold">
            GestorBuy
          </Link>
          <nav className="flex items-center gap-4 text-sm text-gray-600">
            <Link href="/products" className="hover:text-gray-900">
              Produtos
            </Link>
            <Link href="/integrations" className="hover:text-gray-900">
              Integrações
            </Link>
            <Link href="/mining" className="hover:text-gray-900">
              Mineração
            </Link>
          </nav>
        </div>
        <LogoutButton />
      </div>
    </header>
  );
}
