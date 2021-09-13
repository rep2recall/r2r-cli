import platform
import os
import subprocess

from ..dir import get_path


def locate_browser(b: str) -> str:
    if b in ["chrome", "google-chrome", "chromium"]:
        return locate_chrome()
    elif b in ["edge", "msedge", "microsoft-edge"]:
        return locate_edge()
    elif not b:
        b = locate_chrome()
        if not b:
            b = locate_edge()
        return b

    return get_path(b)


def locate_chrome() -> str:
    paths: list[str] = [
        "/usr/bin/google-chrome-stable",
        "/usr/bin/google-chrome",
        "/usr/bin/chromium",
        "/usr/bin/chromium-browser",
        "/snap/bin/chromium",
        "/usr/bin/ungoogled-chromium",
        "/usr/bin/ungoogled-chromium-browser",
    ]

    if platform.system() == "Darwin":
        paths = [
            "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
            "/Applications/Google Chrome Canary.app/Contents/MacOS/Google Chrome Canary",
            "/Applications/Chromium.app/Contents/MacOS/Chromium",
            "/usr/bin/google-chrome-stable",
            "/usr/bin/google-chrome",
            "/usr/bin/chromium",
            "/usr/bin/chromium-browser",
            "/usr/bin/ungoogled-chromium",
            "/usr/bin/ungoogled-chromium-browser",
        ]
    elif platform.system() == "Windows":
        paths = [
            os.getenv("LocalAppData") + "/Google/Chrome/Application/chrome.exe",
            os.getenv("ProgramFiles") + "/Google/Chrome/Application/chrome.exe",
            os.getenv("ProgramFiles(x86)") + "/Google/Chrome/Application/chrome.exe",
            os.getenv("LocalAppData") + "/Chromium/Application/chrome.exe",
            os.getenv("ProgramFiles") + "/Chromium/Application/chrome.exe",
            os.getenv("ProgramFiles(x86)") + "/Chromium/Application/chrome.exe",
        ]

    for p in paths:
        if os.path.isfile(p):
            return p

    paths = []
    if platform.system() == "Linux":
        flatpak_set = set(
            subprocess.check_output(["flatpak", "list", "--columns=application"])
            .decode()
            .split("\n")
        )

        if len(flatpak_set):
            app_ids = [
                "com.google.Chrome",
                "org.chromium.Chromium",
                "com.github.Eloston.UngoogledChromium",
            ]

            for a in app_ids:
                if a in flatpak_set:
                    paths += [
                        f"/var/lib/flatpak/exports/bin/{a}",
                        os.getenv("HOME") + f"/.local/share/flatpak/exports/bin/{a}",
                    ]

    for p in paths:
        if os.path.isfile(p):
            return p

    return ""


def locate_edge() -> str:
    paths: list[str] = [
        "/usr/bin/microsoft-edge",
        "/usr/bin/microsoft-edge-beta",
        "/usr/bin/microsoft-edge-dev",
    ]

    if platform.system() == "Darwin":
        paths = [
            "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
            "/usr/bin/microsoft-edge",
            "/usr/bin/microsoft-edge-beta",
            "/usr/bin/microsoft-edge-dev",
        ]
    elif platform.system() == "Windows":
        paths = [
            os.getenv("ProgramFiles") + "/Microsoft/Edge/Application/msedge.exe",
            os.getenv("ProgramFiles(x86)") + "/Microsoft/Edge/Application/msedge.exe",
        ]

    for p in paths:
        if os.path.isfile(p):
            return p

    return ""
