import { describe, it, expect } from "vitest";
import { http, HttpResponse } from "msw";
import { server } from "./server";

describe("MSW Diagnostics", () => {
  it("intercepts a simple fetch request", async () => {
    server.use(
      http.get("http://localhost:4000/api/test", () => {
        return HttpResponse.json({ ok: true });
      }),
    );

    const response = await fetch("http://localhost:4000/api/test");
    const data = await response.json();

    expect(data).toEqual({ ok: true });
  });
});
