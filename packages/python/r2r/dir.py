import sys
from os import path


def is_pyinstaller() -> bool:
    return getattr(sys, "frozen", False) and hasattr(sys, "_MEIPASS")


def get_path(*s: list[str]) -> str:
    dirname = path.abspath(path.dirname(path.dirname(__file__)))
    if is_pyinstaller():
        dirname = getattr(sys, "_MEIPASS")

    return path.join(dirname, *s)
