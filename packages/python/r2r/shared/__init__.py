import os
import secrets
import yaml

from ..dir import get_external_path

CONFIG_PATH = get_external_path("config.yaml")
config = {}

if os.path.isfile(CONFIG_PATH):
    with open(CONFIG_PATH, "r") as f:
        config = yaml.safe_load(f)

if not config.get("db"):
    config["db"] = "data.db"

if not config.get("port"):
    config["port"] = 25459

if not config.get("secret"):
    config["secret"] = secrets.token_urlsafe(32)

with open(CONFIG_PATH, "w") as f:
    yaml.safe_dump(config, f)
