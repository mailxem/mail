import { GET as getAnalytics } from "../v2/[resource]/route";
export async function GET(request: Request) {
  return getAnalytics(request, {
    params: Promise.resolve({ resource: "report" }),
  });
}
