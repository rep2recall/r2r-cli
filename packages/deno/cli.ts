import { argParse, expandGlob, pathResolve } from "./deps.ts";
import { doQuiz, getIdsFromPath } from "./api/quiz.ts";
import { LocalDB } from "./db/index.ts";
import { runServer } from "./server.ts";

for await (const file of expandGlob(pathResolve("plugins/*.{ts,js}"))) {
  await import(file.path);
}

const args = argParse(Deno.args);
const db = new LocalDB();

switch (args._[0]) {
  case "load":
    for (const p of getPathsFromArgs()) {
      await db.doLoad(p);
    }
    break;
  case "clean":
    db.clean();
    break;
  case "quiz":
    await (async () => {
      const filter = String(args.filter || "");
      let ids: string[] | undefined = undefined;

      if (args._.length > 0) {
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

db.close();

function getPathsFromArgs(): string[] {
  const paths = args._.slice(1).map(String);
  if (!paths.length) {
    throw new Error("There must be paths");
  }
  return paths;
}
