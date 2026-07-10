import assert from "node:assert/strict";
import test from "node:test";
import { resultByLabel, runFixture } from "../helpers/runner.ts";

test("updates route ranking when congestion changes", () => {
  const result = runFixture("congestion_changes");
  const before = resultByLabel(result, "before").quotes ?? [];
  const after = resultByLabel(result, "after").quotes ?? [];
  const route = resultByLabel(result, "north-congested").route;

  assert.equal(before[0].routeId, "route:north-prime");
  assert.equal(route?.congestionBps, 3500);
  assert.equal(after[0].routeId, "route:aurora-rfq");
  assert.ok(
    (after.find((quote) => quote.routeId === "route:north-prime")?.netOut ?? 0) < after[0].netOut,
  );
  assert.equal(result.snapshot.auditIssues.length, 0);
});
