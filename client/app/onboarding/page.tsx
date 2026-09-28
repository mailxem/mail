import { auth } from "@/auth";
import { redirect } from "next/navigation";
import { Onboarding } from "@/components/sending/onboarding";

export default async function Page() {
  const session = await auth();
  if (!session?.user) redirect("/auth/login");
  if (!["ADMIN", "SUPER_ADMIN"].includes(session.user.role || ""))
    redirect("/");
  return <Onboarding />;
}
