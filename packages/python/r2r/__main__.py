import click
import toml

from .dir import get_path

pkg = toml.load(get_path("pyproject.toml"))


@click.command("r2r")
@click.version_option(pkg["tool"]["poetry"]["version"])
def main():
    """Repeat Until Recall - a simple, yet powerful, flashcard app"""
    print("hello")


if __name__ == "__main__":
    main()
