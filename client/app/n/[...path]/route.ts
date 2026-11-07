// Newsletter archives moved to their websites at xem.email/{slug}; old links
// (pages and the RSS feed) redirect there for good.
const site = (process.env.NEXT_PUBLIC_SITE_URL || "https://xem.email").replace(
  /\/$/,
  "",
);

export async function GET(
  _: Request,
  { params }: { params: Promise<{ path: string[] }> },
) {
  const { path } = await params;
  return Response.redirect(
    `${site}/${path.map(encodeURIComponent).join("/")}`,
    308,
  );
}
