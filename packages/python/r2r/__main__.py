import click
import toml

from .dir import get_internal_path
from .shared import config

pkg = toml.load(get_internal_path("pyproject.toml"))


@click.group(invoke_without_command=True)
@click.option("-o", "--db", help=f"database to use  [default: {config['db']}]")
@click.option(
    "-p",
    "--port",
    type=int,
    help=f"port to run the server  [default: {config['port']}]",
)
@click.option(
    "--browser",
    help="browser to open  [default: Chrome with Edge fallback]",
)
@click.option(
    "--mode",
    type=click.Choice(["server", "proxy", "quiz"]),
    help="special mode to run in (server / proxy / quiz)  [optional]",
)
@click.option(
    "-f",
    "--file",
    multiple=True,
    help="files to use (must be loaded first)  [optional]",
)
@click.version_option(pkg["tool"]["poetry"]["version"])
@click.pass_context
def main(
    ctx=None,
    db: str = config["db"],
    port: int = config["port"],
    browser: str = None,
    mode: str = None,
    file: list[str] = list(),
):
    """Repeat Until Recall - a simple, yet powerful, flashcard app"""

    ctx.ensure_object(dict)

    if ctx.invoked_subcommand is None:
        print(db)

    ctx.obj["db"] = db
    ctx.obj["port"] = port
    ctx.obj["file"] = file


@main.command()
@click.option("-o", "--db", help=f"database to use  [default: {config['db']}]")
@click.option(
    "-p",
    "--port",
    type=int,
    help=f"port to run the server  [default: {config['port']}]",
)
@click.option(
    "--debug",
    type=bool,
    help="open the browser as data is being loaded to the database  [optional]",
)
@click.argument("files", nargs=-1)
@click.pass_context
def load(
    ctx=None,
    db: str = None,
    port: int = None,
    debug: bool = None,
    files: list[str] = list(),
):
    """load YAML into the database and exit"""

    if not db:
        db = ctx.obj["db"]

    if not port:
        port = ctx.obj["port"]

    if not len(files):
        files = ctx.obj["file"]

    if not len(files):
        raise KeyError("Number of files to be loaded must not be empty")


if __name__ == "__main__":
    main(obj={})
