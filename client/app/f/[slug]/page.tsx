import { PublicForm } from "@/components/marketing/public-form";
export default async function FormPage({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  return <PublicForm slug={slug} />;
}
