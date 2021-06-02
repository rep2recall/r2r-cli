import { argParse } from "./deps.ts";
import { doLoad } from "./api/load.ts";
import { doQuiz, getIdsFromPath } from "./api/quiz.ts";
import { LocalDB } from "./db/index.ts";
import { runServer } from "./server.ts";

const args = argParse(Deno.args);
const db = new LocalDB();

switch (args._[0]) {
  case "load":
    for (const p of getPathsFromArgs()) {
      await doLoad(p);
    }
    break;
  case "quiz":
    await (async () => {
      const filter = String(args.filter || "");
      let ids: string[] | undefined = undefined;

      if (args._.length > 1) {
        ids = (await Promise.all(
          getPathsFromArgs().map((p) => getIdsFromPath(p)),
        ))
          .flat();
      }

      const sessionId = doQuiz(filter, ids);
      await runServer().then((s) => s.open(`/quiz?sessionId=${sessionId}`))
        .then((s) => s ? s.close() : null);
    })();
    break;
  default:
    await runServer().then((s) => s.open())
      .then((s) => s ? s.close() : null);
}

function getPathsFromArgs(): string[] {
  const paths = args._.slice(1).map(String);
  if (!paths.length) {
    throw new Error("There must be paths");
  }
  return paths;
}
