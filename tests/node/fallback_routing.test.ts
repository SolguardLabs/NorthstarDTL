import assert from "node:assert/strict";
import test from "node:test";
import { resultByLabel, runFixture } from "../helpers/runner.ts";

test("executes ready tickets on an available fallback route", () => {
  const result = runFixture("fallback_routing");
  const submit = resultByLabel(result, "admit").submit;
  const execution = resultByLabel(result, "execute").execute;

  assert.equal(submit?.ticket.plannedRouteId, "route:north-prime");
  assert.equal(execution?.receipts.length, 1);
  assert.equal(execution?.receipts[0].usedFallback, true);
  assert.equal(execution?.receipts[0].routeId, "route:aurora-rfq");
  assert.equal(result.snapshot.queue.length, 0);
  assert.equal(result.snapshot.auditIssues.length, 0);
});
