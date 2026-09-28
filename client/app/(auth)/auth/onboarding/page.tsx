import { redirect } from "next/navigation";

// Registration already creates the workspace. Keep old onboarding URLs valid
// without creating a second team or asking users to repeat setup.
export default function LegacyOnboardingPage() {
  redirect("/onboarding");
}
