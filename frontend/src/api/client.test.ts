import { describe, it, expect } from "vitest";

import { apiUrl } from "./client";

describe("apiUrl", () => {
    it("prefixes the api origin (empty on web, so same-origin relative)", () => {
        expect(apiUrl("/api/v1/site-info")).toBe("/api/v1/site-info");
    });
});
