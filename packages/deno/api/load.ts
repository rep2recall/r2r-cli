import { pathResolve, yamlParse, z } from "../deps.ts";

const zData = z.object({
  model: z.array(z.object({
    _id: z.string(),
    name: z.string().optional(),
    front: z.string().optional(),
    back: z.string().optional(),
    shared: z.string().optional(),
    generated: z.object({
      _: z.string().optional(),
    }).nonstrict(),
  })),
  template: z.array(z.object({
    _id: z.string(),
    model: z.string().optional(),
    name: z.string().optional(),
    front: z.string().optional(),
    back: z.string().optional(),
    shared: z.string().optional(),
  })),
  note: z.array(z.object({
    _id: z.string(),
    model: z.string().optional(),
    data: z.object({}).nonstrict(),
  })),
});

export async function doLoad(filename: string) {
  const dataFile = await zData.parseAsync(yamlParse(
    await Deno.readTextFile(pathResolve(filename)),
  ));
}
