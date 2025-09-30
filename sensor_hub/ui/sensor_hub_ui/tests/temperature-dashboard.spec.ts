import { test, expect } from "@playwright/test";

test.beforeEach(async ({ page }) => {
  await page.goto("http://localhost:3000/");
});

test.describe(
  "Temperature Dashboard - Rendering of components",
  { tag: "@visual" },
  () => {
    test("should display the correct title", async ({ page }) => {
      await expect(page.locator("h1")).toHaveText("Temperature Dashboard");
    });

    test("should show current temperature card", async ({ page }) => {
      const currentTempCards = await page.getByTestId(
        "current-temperature-card"
      );
      currentTempCards.first();
      await expect(currentTempCards.first()).toBeVisible();
    });

    test("should have button to collect sensor data", async ({ page }) => {
      const sensorButtons = page.getByRole("button", { name: /Trigger/i });
      await expect(sensorButtons.first()).toBeVisible();
    });

    test("should display temperature graph", async ({ page }) => {
      await expect(
        page.locator("h2", { hasText: "Temperature Over Time" })
      ).toBeVisible();
      await expect(page.getByTestId("temperature-graph")).toBeVisible();
    });

    test("should have date range picker for start and end dates", async ({
      page,
    }) => {
      const startDatePicker = page.getByLabel("Start Date");
      const endDatePicker = page.getByLabel("End Date");
      await expect(startDatePicker.first()).toBeVisible();
      await expect(endDatePicker.first()).toBeVisible();
    });

    test("should have toggle for hourly averages", async ({ page }) => {
      const toggle = page.getByRole("checkbox", { name: /Hourly Averages/i });
      await expect(toggle).toBeVisible();
    });
  }
);

test.describe(
  "Temperature Dashboard - Interactivity",
  { tag: "@interactive" },
  () => {
    test("should trigger sensor reading on button click", async ({ page }) => {
      const sensorButton = page
        .getByRole("button", { name: /Trigger/i })
        .first();

      const [request] = await Promise.all([
        page.waitForRequest(
          (req) =>
            req.url().includes("/sensors/temperature") && req.method() === "GET"
        ),
        sensorButton.click(),
      ]);

      expect(request).toBeTruthy();
    });

    test("should show invalid date range error", async ({ page }) => {
      const startDatePicker = page.getByTestId("start-date-picker");

      const startPickerButton = startDatePicker.locator("button");
      await startPickerButton.click();

      const startDateCalendarDialog = page.getByRole("dialog");
      await expect(startDateCalendarDialog).toBeVisible();

      const tomorrow = new Date();
      tomorrow.setDate(tomorrow.getDate() + 1);
      const currentEndDay = tomorrow.getDate();
      const currentMonth = tomorrow.toLocaleString("default", {
        month: "long",
      });
      const currentYear = tomorrow.getFullYear();

      if (currentEndDay - 7 < 1) {
        // if the current day is less than 7, then the calendar will open to the previous month
        const previousMonth = new Date();
        previousMonth.setMonth(tomorrow.getMonth() - 1);
        const prevMonthName = previousMonth.toLocaleString("default", {
          month: "long",
        });
        const prevMonthYear = previousMonth.getFullYear();
        await expect(
          startDateCalendarDialog.getByText(`${prevMonthName} ${prevMonthYear}`)
        ).toBeVisible();
        // need to move to the next month to find the current month

        await expect(startDateCalendarDialog).toBeVisible();

        // this is flaky on chromium without this tightrope of delays and timeouts
        // seems to be ok on firefox and webkit
        await page.waitForTimeout(200);
        await startDateCalendarDialog
          .getByRole("button", { name: "Next month" })
          .click({ trial: false, delay: 100 });
        await expect(startDateCalendarDialog).toBeVisible({ timeout: 2000 });

        await expect(
          startDateCalendarDialog.getByText(`${currentMonth} ${currentYear}`)
        ).toBeVisible();
        await expect(
          startDateCalendarDialog.getByText(`${prevMonthName} ${prevMonthYear}`)
        ).toBeHidden({ timeout: 2000 });
      }

      const dayToClick = (currentEndDay + 1).toString();

      // have to pick the first after scrolling months to the correct month
      // there seems to be remaining instances of the previous month in the DOM
      await startDateCalendarDialog
        .getByRole("gridcell", { name: dayToClick, exact: true })
        .first()
        .click();

      await startDateCalendarDialog.waitFor({ state: "hidden" });

      const errorMessage = page.getByText("Invalid date range");
      await expect(errorMessage).toBeVisible();
    });

    test("should update graph when date range is changed", async ({ page }) => {
      const startDatePicker = page.getByTestId("start-date-picker");
      const endDatePicker = page.getByTestId("end-date-picker");

      const startPickerButton = startDatePicker.locator("button");
      const endPickerButton = endDatePicker.locator("button");
      await startPickerButton.click();
      const startDateCalendarDialog = page.locator('[role="dialog"]');

      await startDateCalendarDialog
        .getByRole("gridcell", { name: "14" })
        .click();

      await startDateCalendarDialog.waitFor({ state: "hidden" });

      await endPickerButton.click();

      const endDateCalendarDialog = page.locator('[role="dialog"]');

      const request = await Promise.all([
        page.waitForRequest(
          (req) =>
            req.url().includes("readings/hourly/between") &&
            req.method() === "GET"
        ),
        await endDateCalendarDialog
          .getByRole("gridcell", { name: "19" })
          .click(),
      ]);

      expect(request).toBeTruthy();
    });

    test("should toggle hourly averages and update graph", async ({ page }) => {
      const toggle = page.getByRole("checkbox", { name: /Hourly Averages/i });

      const request = await Promise.all([
        page.waitForRequest(
          (req) =>
            req.url().includes("readings/between") && req.method() === "GET"
        ),
        toggle.click(),
      ]);

      expect(request).toBeTruthy();
    });
  }
);
