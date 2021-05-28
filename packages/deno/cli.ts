import { serve } from "https://deno.land/std@0.97.0/http/server.ts";
import {
  parse as yamlParse,
} from "https://deno.land/std@0.82.0/encoding/yaml.ts";
import { locateChrome } from "./app.ts";

const cfg = yamlParse(await Deno.readTextFile("config.yaml")) as {
  port: number;
  app?: {
    width?: number;
    height?: number;
    chromePath?: string;
    remoteDebuggingPort?: number;
  };
};

if (cfg.app) {
  cfg.app.chromePath = await locateChrome(cfg.app.chromePath) || undefined;
}

const s = serve({ port: cfg.port });
console.log(`Server is listening at http://localhost:${cfg.port}`);

if (cfg.app?.chromePath) {
  let remoteDebuggingPort = cfg.app.remoteDebuggingPort;
  if (!remoteDebuggingPort) {
    const s1 = serve({ port: 0 });
    remoteDebuggingPort = (s1.listener.addr as Deno.NetAddr).port;
    s1.close();
  }

  Deno.run({
    cmd: [
      cfg.app.chromePath,
      `--app=http://localhost:${cfg.port}`,
      `--window-size=${cfg.app.width || 800},${cfg.app.height || 600}`,
      `--remote-debugging-port=${remoteDebuggingPort}`,
    ],
    stdout: "inherit",
    stdin: "inherit",
  });

  (async () => {
    await new Promise((resolve) => setTimeout(resolve, 5000));

    while (true) {
      await new Promise((resolve) => setTimeout(resolve, 1000));

      try {
        const s1 = serve({ port: remoteDebuggingPort });
        s1.close();
        break;
      } catch (e) {
        if (!(e instanceof Deno.errors.AddrInUse)) {
          throw e;
        }
      }
    }

    s.close();
  })();
}

for await (const req of s) {
  req.respond({ body: "Hello World\n" });
}
