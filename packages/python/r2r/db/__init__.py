from datetime import datetime
import json
import functools
from contextvars import ContextVar
import peewee as pw
from playhouse import sqlite_ext

from .quiz import srs_map, get_next_review
from ..dir import get_external_path
from ..shared import config

db_state_default = {"closed": None, "conn": None, "ctx": None, "transactions": None}
db_state = ContextVar("db_state", default=db_state_default.copy())


class PeeweeConnectionState(pw._ConnectionState):
    def __init__(self, **kwargs):
        super().__setattr__("_state", db_state)
        super().__init__(**kwargs)

    def __setattr__(self, name, value):
        self._state.get()[name] = value

    def __getattr__(self, name):
        return self._state.get()[name]


db = sqlite_ext.SqliteExtDatabase(
    get_external_path(config["db"]),
    check_same_thread=False,
    pragmas={"journal_mode": "wal", "cache_size": 10000, "foreign_keys": 1},
)
db._state = PeeweeConnectionState()


@db.func()
def tokenize(value: str, lang: str):
    return value


class BaseModel(pw.Model):
    class Meta:
        database = db


class TimestampModel(BaseModel):
    created_at = pw.DateTimeField(
        formats="%Y-%m-%d %H:%M:%S", constraints=[pw.SQL("DEFAULT (datetime('now'))")]
    )
    updated_at = pw.DateTimeField(
        formats="%Y-%m-%d %H:%M:%S", constraints=[pw.SQL("DEFAULT (datetime('now'))")]
    )

    @classmethod
    def create_table(cls, *args, **kwargs):
        super().create_table(*args, **kwargs)
        cls._meta.database.execute_sql(
            f"""
            CREATE TRIGGER IF NOT EXISTS t_{cls._meta.name}_updated_at AFTER UPDATE ON {cls._meta.name}
            WHEN OLD.created_at = NEW.created_at AND OLD.updated_at = NEW.updated_at
            BEGIN
                UPDATE ${cls._meta.name} SET updated_at = datetime('now') WHERE ROWID = NEW.ROWID;
            END;
            """
        )


class TagField(pw.TextField):
    def db_value(self, value: str):
        if value and value[0] == " " and value[-1] == " " and value[1]:
            return value[1:-1].split(" ")
        return []

    def python_value(self, value: list[str]):
        if len(value) > 0:
            return " " + " ".join(value) + " "
        return ""


class Model(TimestampModel):
    name = pw.CharField(null=True, index=True)
    front = pw.CharField(default="")
    back = pw.CharField(default="")
    shared = pw.CharField(default="")
    generator = sqlite_ext.JSONField(
        json_dumps=functools.partial(json.dumps, ensure_ascii=False)
    )


class Template(TimestampModel):
    model = pw.ForeignKeyField(
        Model, backref="templates", on_delete="CASCADE", index=True
    )
    name = pw.CharField(null=True, index=True)
    front = pw.CharField(default="")
    back = pw.CharField(default="")
    shared = pw.CharField(default="")
    only_if = pw.CharField(default="")


class Note(TimestampModel):
    model = pw.ForeignKeyField(Model, backref="notes", on_delete="CASCADE", index=True)


class Attr(TimestampModel):
    note = pw.ForeignKeyField(
        Note, field="note_id", backref="attrs", on_delete="CASCADE", index=True
    )
    key = pw.CharField()
    value = sqlite_ext.JSONField(
        json_dumps=functools.partial(json.dumps, ensure_ascii=False)
    )
    lang = pw.CharField(null=True)

    @classmethod
    def create_table(cls, *args, **kwargs):
        super().create_table(*args, **kwargs)
        cls._meta.database.execute_sql(
            f"""
            CREATE VIRTUAL TABLE IF NOT EXISTS note_fts USING fts5(
                note_id UNINDEXED,
                key,
                value,
                content=note_attr,
                content_rowid=id,
                tokenize=porter
            );

            -- Triggers to keep the FTS index up to date.
            CREATE TRIGGER IF NOT EXISTS t_note_attr_ai AFTER INSERT ON note_attr BEGIN
                INSERT INTO note_fts(rowid, note_id, key, value) VALUES (new.id, new.note_id, new.key, tokenize(new.value, new.lang));
            END;
            CREATE TRIGGER IF NOT EXISTS t_note_attr_ad AFTER DELETE ON note_attr BEGIN
                INSERT INTO note_fts(note_fts, rowid, note_id, key, value) VALUES ('delete', old.id, old.note_id, old.key, tokenize(old.value, old.lang));
            END;
            CREATE TRIGGER IF NOT EXISTS t_note_attr_au AFTER UPDATE ON note_attr BEGIN
                INSERT INTO note_fts(note_fts, rowid, note_id, key, value) VALUES ('delete', old.id, old.note_id, old.key, tokenize(old.value, old.lang));
                INSERT INTO note_fts(rowid, note_id, key, value) VALUES (new.id, new.note_id, new.key, tokenize(new.value, new.lang));
            END;
            """
        )


class Card(TimestampModel):
    template = pw.ForeignKeyField(Template, backref="cards", null=True)
    note = pw.ForeignKeyField(Note, backref="cards", null=True)
    front = pw.CharField(default="")
    back = pw.CharField(default="")
    shared = pw.CharField(default="")
    mnemonic = pw.CharField(default="")
    srs_level = pw.IntegerField(null=True, index=True)
    next_review = pw.DateTimeField(null=True, index=True)
    last_right = pw.DateTimeField(null=True, index=True)
    last_wrong = pw.DateTimeField(null=True, index=True)
    max_right = pw.IntegerField(default=0, index=True)
    max_wrong = pw.IntegerField(default=0, index=True)
    right_streak = pw.IntegerField(default=0, index=True)
    wrong_streak = pw.IntegerField(default=0, index=True)
    tag = TagField(null=True, index=True)

    def update_srs_level(self, d: int) -> None:
        now = datetime.now()

        if d > 0:
            self.last_right = now
            self.right_streak = self.right_streak + 1

            if self.right_streak > self.max_right:
                self.max_right = self.right_streak
        elif d < 0:
            self.last_wrong = now
            self.wrong_streak = self.wrong_streak + 1

            if self.wrong_streak > self.max_wrong:
                self.max_wrong = self.wrong_streak

        self.srs_level = self.srs_level + d if self.srs_level else d

        if self.srs_level >= len(srs_map):
            self.srs_level = len(srs_map) - 1

        if self.srs_level < 0:
            self.srs_level = 0
            self.next_review = get_next_review()
        else:
            self.next_review = get_next_review(self.srs_level)


def create_tables():
    db.connect()
    db.create_tables([Model, Template, Note, Attr, Card])
    db.close()
