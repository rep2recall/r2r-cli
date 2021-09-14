import sys
from os import path


def is_pyinstaller() -> bool:
    return getattr(sys, "frozen", False) and hasattr(sys, "_MEIPASS")


def get_internal_path(*s: list[str]) -> str:
    return path.join(
        getattr(sys, "_MEIPASS")
        if is_pyinstaller()
        else path.abspath(path.dirname(path.dirname(__file__))),
        *s
    )


def get_external_path(*s: list[str]) -> str:
    return (
        path.join(path.dirname(sys.executable), *s)
        if is_pyinstaller()
        else get_internal_path(*s)
    )
