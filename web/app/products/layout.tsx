import Link from "next/link";
import { LogoutButton } from "./logout-button";

export default function ProductsLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <div className="min-h-screen bg-gray-50">
      <header className="border-b border-gray-200 bg-white">
        <div className="mx-auto flex max-w-4xl items-center justify-between px-4 py-4">
          <Link href="/products" className="font-semibold">
            GestorBuy
          </Link>
          <LogoutButton />
        </div>
      </header>
      {children}
    </div>
  );
}
