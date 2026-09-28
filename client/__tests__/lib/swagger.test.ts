import { baseApiSpec, createRouteSpec } from "@/app/lib/swagger";
import { describe, expect, it } from "@jest/globals";
describe("API documentation", () => {
  it("declares bearer authentication", () => {
    expect(baseApiSpec.openapi).toBe("3.0.0");
    expect(baseApiSpec.components?.securitySchemes?.bearerAuth).toMatchObject({
      type: "http",
      scheme: "bearer",
    });
  });
  it("preserves operation responses and documents authorization errors", () => {
    const route = createRouteSpec("/forms", {
      get: { responses: { 200: { description: "Workspace forms" } } },
    });
    expect(route.get?.responses[200]).toEqual({
      description: "Workspace forms",
    });
    expect(route.get?.responses[401]).toBeDefined();
    expect(route.get?.responses[500]).toBeDefined();
  });
});
