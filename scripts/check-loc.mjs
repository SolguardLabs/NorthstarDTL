import { readdirSync, readFileSync, statSync } from "node:fs";
import { join } from "node:path";

const root = join(process.cwd(), "src");
const min = 3000;
const max = 4000;

function files(dir) {
  return readdirSync(dir)
    .map((name) => join(dir, name))
    .flatMap((path) => (statSync(path).isDirectory() ? files(path) : [path]));
}

const total = files(root)
  .filter((path) => path.endsWith(".go"))
  .map((path) => readFileSync(path, "utf8").split(/\r?\n/).filter(Boolean).length)
  .reduce((sum, count) => sum + count, 0);

console.log(`src LOC: ${total}`);
if (total < min || total > max) {
  console.error(`expected src LOC between ${min} and ${max}`);
  process.exit(1);
}
