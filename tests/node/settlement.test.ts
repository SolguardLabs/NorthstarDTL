import assert from "node:assert/strict";
import test from "node:test";
import { balanceOf, resultByLabel, runFixture } from "../helpers/runner.ts";

test("settles a ready ticket and posts source, destination and fee balances", () => {
  const result = runFixture("settlement");
  const execution = resultByLabel(result, "execute").execute;

  assert.equal(execution?.receipts.length, 1);
  assert.equal(execution?.receipts[0].routeId, "route:north-prime");
  assert.equal(execution?.receipts[0].amountIn, 100000);
  assert.equal(execution?.receipts[0].destinationAmount, 101000);
  assert.equal(execution?.receipts[0].routeFee, 150);

  assert.equal(balanceOf(result, "acct:alice", "usdc").available, 1399850);
  assert.equal(balanceOf(result, "acct:alice", "usdc").reserved, 0);
  assert.equal(balanceOf(result, "acct:bob", "eurc").available, 101000);
  assert.equal(balanceOf(result, "acct:north-prime-usdc", "usdc").available, 100000);
  assert.equal(balanceOf(result, "acct:fees", "usdc").available, 150);

  const route = result.snapshot.routes.find((entry) => entry.id === "route:north-prime");
  assert.equal(route?.exposure, 100000);
  assert.equal(route?.outputLiquidity, 1899000);
  assert.equal(result.snapshot.auditIssues.length, 0);
});
