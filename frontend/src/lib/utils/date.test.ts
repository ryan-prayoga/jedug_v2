import { describe, it, expect } from "vitest";
import { relativeTime, relativeTimeLabel } from "./date";

describe("relativeTime", () => {
	it("returns 'baru saja' for timestamps less than a minute ago", () => {
		expect(relativeTime(new Date().toISOString())).toBe("baru saja");
	});

	it("returns minutes ago for timestamps within an hour", () => {
		const fiveMinAgo = new Date(Date.now() - 5 * 60 * 1000).toISOString();
		expect(relativeTime(fiveMinAgo)).toBe("5 menit lalu");
	});

	it("returns hours ago for timestamps within a day", () => {
		const threeHoursAgo = new Date(Date.now() - 3 * 60 * 60 * 1000).toISOString();
		expect(relativeTime(threeHoursAgo)).toBe("3 jam lalu");
	});

	it("returns days ago for timestamps within a month", () => {
		const fourDaysAgo = new Date(Date.now() - 4 * 24 * 60 * 60 * 1000).toISOString();
		expect(relativeTime(fourDaysAgo)).toBe("4 hari lalu");
	});
});

describe("relativeTimeLabel", () => {
	it("capitalizes the first character of the relative time", () => {
		expect(relativeTimeLabel(new Date().toISOString())).toBe("Baru saja");
	});
});
