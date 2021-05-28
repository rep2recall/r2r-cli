export async function locateChrome(chrome?: string) {
  if (chrome) {
    try {
      await Deno.stat(chrome);
      return chrome;
    } catch (e) {
      if (!(e instanceof Deno.errors.NotFound)) {
        throw e;
      }
    }
  }

  switch (Deno.build.os) {
    case "windows":
      for (
        const p of [
          Deno.env.get("LocalAppData") +
          "/Google/Chrome/Application/chrome.exe",
          Deno.env.get("ProgramFiles") +
          "/Google/Chrome/Application/chrome.exe",
          Deno.env.get("ProgramFiles(x86)") +
          "/Google/Chrome/Application/chrome.exe",
          Deno.env.get("LocalAppData") + "/Chromium/Application/chrome.exe",
          Deno.env.get("ProgramFiles") + "/Chromium/Application/chrome.exe",
          Deno.env.get("ProgramFiles(x86)") +
          "/Chromium/Application/chrome.exe",
          Deno.env.get("ProgramFiles(x86)") +
          "/Microsoft/Edge/Application/msedge.exe",
        ]
      ) {
        try {
          await Deno.stat(p);
          return p;
        } catch (e) {
          if (!(e instanceof Deno.errors.NotFound)) {
            throw e;
          }
        }
      }
      break;
    case "darwin":
      for (
        const p of [
          "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
          "/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
          "/Applications/Chromium.app/Contents/MacOS/Chromium",
          "/usr/bin/google-chrome-stable",
          "/usr/bin/google-chrome",
          "/usr/bin/chromium",
          "/usr/bin/chromium-browser",
        ]
      ) {
        try {
          await Deno.stat(p);
          return p;
        } catch (e) {
          if (!(e instanceof Deno.errors.NotFound)) {
            throw e;
          }
        }
      }
      break;
    default:
      for (
        const p of [
          "/usr/bin/google-chrome-stable",
          "/usr/bin/google-chrome",
          "/usr/bin/chromium",
          "/usr/bin/chromium-browser",
          "/snap/bin/chromium",
        ]
      ) {
        try {
          await Deno.stat(p);
          return p;
        } catch (e) {
          if (!(e instanceof Deno.errors.NotFound)) {
            throw e;
          }
        }
      }
  }

  return null;
}
