import { spawnSync } from "node:child_process";
import { dirname, join } from "node:path";
import { fileURLToPath } from "node:url";

const here = dirname(fileURLToPath(import.meta.url));
export const projectRoot = join(here, "..", "..");

export type Quote = {
  routeId: string;
  amountIn: number;
  grossOut: number;
  congestionPenalty: number;
  routeFee: number;
  netOut: number;
  score: number;
  readyEpoch: number;
  expiresEpoch: number;
  reason?: string;
};

export type ScenarioResult = {
  name: string;
  results: Array<{
    type: string;
    label: string;
    quotes?: Quote[];
    submit?: {
      ticket: {
        id: string;
        plannedRouteId: string;
        activeRouteId: string;
        readyEpoch: number;
        quote: Quote;
      };
      quote: Quote;
    };
    execute?: {
      receipts: Array<{
        id: string;
        ticketId: string;
        intentId: string;
        routeId: string;
        quotedRouteId: string;
        amountIn: number;
        destinationAmount: number;
        routeFee: number;
        usedFallback: boolean;
      }>;
      deferred: unknown[];
    };
    epoch?: number;
    route?: {
      id: string;
      status: string;
      congestionBps: number;
      congestionLevel: number;
    };
    snapshot?: ScenarioResult["snapshot"];
  }>;
  snapshot: {
    epoch: number;
    balances: Array<{ account: string; asset: string; available: number; reserved: number }>;
    routes: Array<{
      id: string;
      ratePpm: number;
      congestionBps: number;
      congestionLevel: number;
      exposure: number;
      outputLiquidity: number;
      status: string;
    }>;
    queue: Array<{ ticketId: string; plannedRouteId: string; activeRouteId: string }>;
    receipts: Array<{
      id: string;
      routeId: string;
      quotedRouteId: string;
      destinationAmount: number;
      routeFee: number;
      usedFallback: boolean;
    }>;
    events: Array<{ type: string; routeId?: string; ticketId?: string; message?: string }>;
    auditIssues: Array<{ code: string; severity: string; message: string }>;
  };
};

export function runFixture(name: string): ScenarioResult {
  const fixturePath = join(projectRoot, "tests", "fixtures", `${name}.json`);
  const child = spawnSync("go", ["run", "./cmd/northstardtl", "run", fixturePath], {
    cwd: projectRoot,
    encoding: "utf8",
  });
  if (child.status !== 0) {
    throw new Error(
      [
        `fixture ${name} failed`,
        `status: ${child.status}`,
        `stdout: ${child.stdout}`,
        `stderr: ${child.stderr}`,
      ].join("\n"),
    );
  }
  return JSON.parse(child.stdout) as ScenarioResult;
}

export function resultByLabel(
  result: ScenarioResult,
  label: string,
): NonNullable<ScenarioResult["results"][number]> {
  const found = result.results.find((entry) => entry.label === label);
  if (!found) {
    throw new Error(`missing action label ${label}`);
  }
  return found;
}

export function balanceOf(
  result: ScenarioResult,
  account: string,
  asset: string,
): { available: number; reserved: number } {
  return (
    result.snapshot.balances.find(
      (balance) => balance.account === account && balance.asset === asset,
    ) ?? { available: 0, reserved: 0 }
  );
}
