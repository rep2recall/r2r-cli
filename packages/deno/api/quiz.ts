import { pathResolve, yamlParse, z } from "../deps.ts";

export function doQuiz(_filter: string, _ids?: string[]): string {
  const sessionId = "";
  return sessionId;
}

export async function getIdsFromPath(filename: string) {
  const contents = await z.array(
    z.object({
      id: z.string(),
    }).nonstrict(),
  ).parseAsync(yamlParse(await Deno.readTextFile(pathResolve(filename))));

  return contents.map((c) => c.id);
}
