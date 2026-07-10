import assert from "node:assert/strict";
import test from "node:test";
import { resultByLabel, runFixture } from "../helpers/runner.ts";

test("scores preferred liquid routes ahead of congested alternatives", () => {
  const result = runFixture("route_selection");
  const quote = resultByLabel(result, "initial-quote").quotes ?? [];

  assert.equal(quote.length, 4);
  assert.equal(quote[0].routeId, "route:north-prime");
  assert.equal(quote[0].grossOut, 101000);
  assert.equal(quote[0].congestionPenalty, 0);
  assert.equal(quote[0].netOut, 101000);
  assert.ok(quote[0].score > quote[1].score);
  assert.equal(result.snapshot.auditIssues.length, 0);
});
