import { expect, test } from "@playwright/test";

test("shows open games, enters the bot table, and starts the hand", async ({
	page,
}) => {
	await page.goto("/");

	await expect(
		page.getByRole("heading", {
			name: "Find a table, create one, or launch your own bot arena.",
		}),
	).toBeVisible();
	await expect(
		page.locator(".lobby-card").filter({ hasText: "Bot Warmup Table" }),
	).toBeVisible();
	await expect(
		page.locator(".lobby-card").filter({ hasText: "Blinds 10 / 20" }),
	).toBeVisible();

	await page.getByRole("button", { name: "Open Table" }).click();

	await expect(page).toHaveURL(/#\/tables\/default$/);
	await expect(page.locator(".poker-table")).toBeVisible();
	await expect(page.locator(".backend-status")).toHaveClass(/is-collapsed/);
	await page.getByRole("button", { name: "Dev Stats" }).click();
	await expect(
		page
			.locator(".backend-grid div")
			.filter({ has: page.locator("span", { hasText: "Table" }) }),
	).toContainText("default");
	await page.getByRole("button", { name: "Hide Dev Stats" }).click();
	await expect(page.locator(".backend-status")).toHaveClass(/is-collapsed/);

	const availableSeatButton = page
		.getByRole("button", { name: /Take Seat \d+|Join Seat \d+ Next Hand/ })
		.first();
	if (await availableSeatButton.isVisible()) {
		await availableSeatButton.click();
	}

	await page.getByRole("button", { name: "Hand Log" }).click();
	await expect(
		page.getByRole("button", { name: "Hide Hand Log" }),
	).toBeVisible();
	await expect(page.locator("#action-log-list")).toBeVisible();
	await page.getByRole("button", { name: "Hide Hand Log" }).click();
	await expect(page.getByRole("button", { name: "Hand Log" })).toBeVisible();

	await expect(
		page.getByRole("button", { name: /Sit Out|Sit In/ }),
	).toBeVisible({
		timeout: 30_000,
	});
	await expect(page.locator("#action-state")).not.toContainText(
		"Join the game to take the open seat.",
	);
	await expect(page.locator("#btn-fold")).toBeVisible();

	await page.getByRole("button", { name: "Dev Stats" }).click();
	await expect(page.locator(".backend-status")).not.toHaveClass(/is-collapsed/);
});
