import { pathResolve, serve, yamlParse, z } from "./deps.ts";
import { locateChrome } from "./chrome.ts";

const zConfig = z.object({
  port: z.number(),
  app: z.object({
    width: z.number().optional(),
    height: z.number().optional(),
    chromePath: z.string().optional(),
    remoteDebuggingPort: z.number().optional(),
  }).optional(),
});

export async function runServer(): Promise<{
  open: (url?: string) => Promise<
    {
      close: () => Promise<void>;
    } | null
  >;
}> {
  const cfg = await zConfig.parseAsync(yamlParse(
    await Deno.readTextFile(pathResolve("config.yaml")),
  ));

  if (cfg.app) {
    cfg.app.chromePath = await locateChrome(cfg.app.chromePath) || undefined;
  }

  const s = serve({ port: cfg.port });
  (async () => {
    for await (const req of s) {
      req.respond({ body: "Hello World\n" });
    }
  })();

  return {
    open: async (url = "") => {
      console.log(`Please go to http://localhost:${cfg.port}${url}`);

      if (!cfg.app || !cfg.app.chromePath) {
        return null;
      }

      let remoteDebuggingPort = cfg.app.remoteDebuggingPort || 0;
      if (!remoteDebuggingPort) {
        const s1 = serve({ port: 0 });
        remoteDebuggingPort = (s1.listener.addr as Deno.NetAddr).port;
        s1.close();
      }

      Deno.run({
        cmd: [
          cfg.app.chromePath,
          `--app=http://localhost:${cfg.port}${url}`,
          `--window-size=${cfg.app.width || 800},${cfg.app.height || 600}`,
          `--remote-debugging-port=${remoteDebuggingPort}`,
        ],
        stdout: "inherit",
        stdin: "inherit",
      });

      while (true) {
        await new Promise((resolve) => setTimeout(resolve, 1000));

        try {
          serve({ port: remoteDebuggingPort }).close();
        } catch (e) {
          if (!(e instanceof Deno.errors.AddrInUse)) {
            throw e;
          } else {
            break;
          }
        }
      }

      return {
        close: async () => {
          while (true) {
            try {
              const s1 = serve({ port: remoteDebuggingPort });
              s1.close();
              break;
            } catch (e) {
              if (!(e instanceof Deno.errors.AddrInUse)) {
                throw e;
              }
            }

            await new Promise((resolve) => setTimeout(resolve, 1000));
          }

          s.close();
        },
      };
    },
  };
}
