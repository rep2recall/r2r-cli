import {
  dirname,
  fromFileUrl,
  resolve as stdResolve,
} from "https://deno.land/std@0.97.0/path/mod.ts";

export function pathResolve(...paths: string[]) {
  return stdResolve(dirname(fromFileUrl(import.meta.url)), ...paths);
}

export { serve } from "https://deno.land/std@0.97.0/http/server.ts";
export {
  parse as yamlParse,
} from "https://deno.land/std@0.97.0/encoding/yaml.ts";
export { parse as argParse } from "https://deno.land/std@0.97.0/flags/mod.ts";

export * as z from "https://deno.land/x/zod@v3.1.0/mod.ts";
export { DB as Sqlite } from "https://deno.land/x/sqlite@v2.4.2/mod.ts";
