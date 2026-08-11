import { redirect } from "next/navigation";
import { cookies } from "next/headers";
import { SESSION_COOKIE } from "@/lib/session";

export default async function Home() {
  const cookieStore = await cookies();
  redirect(cookieStore.has(SESSION_COOKIE) ? "/products" : "/login");
}
