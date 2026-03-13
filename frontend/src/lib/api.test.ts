import { expect, test } from "vitest";
import { getApiErrorMessage } from "./api";

test("prefers flattened field validation errors", () => {
  expect(
    getApiErrorMessage({
      error: "Please fix the highlighted fields.",
      errors: {
        email: ["must have the @ sign and no spaces"],
        password: ["can't be blank"],
      },
    }),
  ).toBe("email must have the @ sign and no spaces, password can't be blank");
});

test("falls back to the top-level API error", () => {
  expect(
    getApiErrorMessage({
      error: "Invalid email or password",
    }),
  ).toBe("Invalid email or password");
});
