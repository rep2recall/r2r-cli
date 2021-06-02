import { pathResolve, Sqlite } from "../deps.ts";

export class LocalDB {
  db: Sqlite;
  isClosed = false;

  constructor(public path = pathResolve("data.db")) {
    this.db = new Sqlite(pathResolve("data.db"));

    /**
     * model
     */
    this.db.query(/* sql */ `
    CREATE TABLE IF NOT EXISTS model (
      _id           TEXT NOT NULL PRIMARY KEY,
      deletedAt     TIMESTAMP,
      [name]        TEXT,
      front         TEXT,
      back          TEXT,
      shared        TEXT
    );

    CREATE INDEX IF NOT EXISTS idx_model_name ON model ([name]);
    `);

    /**
    * template
    */
    this.db.query(/* sql */ `
    CREATE TABLE IF NOT EXISTS template (
      _id           TEXT NOT NULL PRIMARY KEY,
      deletedAt     TIMESTAMP,
      model         TEXT NOT NULL REFERENCES model(_id) ON DELETE CASCADE,
      [name]        TEXT,  -- should not duplicate, although not restricted; otherwise, name can't be null
      front         TEXT,
      back          TEXT,
      shared        TEXT
    );

    CREATE INDEX IF NOT EXISTS idx_template_model ON template (model);
    CREATE INDEX IF NOT EXISTS idx_template_name ON template ([name]);
    `);

    /**
    * note
    */
    this.db.query(/* sql */ `
    CREATE TABLE IF NOT EXISTS note USING fts5(
      deletedAt     UNINDEXED,  -- TIMESTAMP
      _id,          -- TEXT NOT NULL
      model,        -- TEXT NOT NULL REFERENCES model(_id) ON DELETE CASCADE
      [key],        -- TEXT NOT NULL
      [data],       -- JSON -- must be JSONified text
      [generated]   UNINDEXED
    );
    `);

    /**
    * card
    */
    this.db.query(/* sql */ `
    CREATE TABLE IF NOT EXISTS [card] (
      -- ROWID
      deletedAt     TIMESTAMP,
      template      TEXT REFERENCES template(_id) ON DELETE CASCADE,  -- template and note must exist hand-in-hand
      note          TEXT REFERENCES note(_id) ON DELETE CASCADE,      -- template and note must exist hand-in-hand
      front         TEXT,
      back          TEXT,
      shared        TEXT,
      srsLevel      INT,
      nextReview    TIMESTAMP,
      lastRight     TIMESTAMP,
      lastWrong     TIMESTAMP,
      maxRight      INT,
      maxWrong      INT,
      rightStreak   INT,
      wrongStreak   INT,
      tag           TEXT   -- searchable line of text. maybe begin-end with space and space-separated
    );

    CREATE INDEX IF NOT EXISTS idx_card_template ON [card] (template);
    CREATE INDEX IF NOT EXISTS idx_card_note ON [card] (note);
    CREATE INDEX IF NOT EXISTS idx_card_srsLevel ON [card] (srsLevel);
    CREATE INDEX IF NOT EXISTS idx_card_nextReview ON [card] (nextReview);
    CREATE INDEX IF NOT EXISTS idx_card_lastRight ON [card] (lastRight);
    CREATE INDEX IF NOT EXISTS idx_card_lastWrong ON [card] (lastWrong);
    CREATE INDEX IF NOT EXISTS idx_card_maxRight ON [card] (maxRight);
    CREATE INDEX IF NOT EXISTS idx_card_maxWrong ON [card] (maxWrong);
    CREATE INDEX IF NOT EXISTS idx_card_rightStreak ON [card] (rightStreak);
    CREATE INDEX IF NOT EXISTS idx_card_wrongStreak ON [card] (wrongStreak);
    CREATE INDEX IF NOT EXISTS idx_card_tag ON [card] (tag);
    `);

    addEventListener("unload", () => {
      this.close();
    });
  }

  close() {
    if (!this.isClosed) {
      this.db.close();
    }
  }

  cleanup = {
    model: () => {},
    template: () => {
      this.db.query(
        /* sql */ `
      UPDATE template
      SET deletedAt = ?
      WHERE
        ([name] IS NOT NULL AND ROWID NOT IN (SELECT ROWID FROM template WHERE [name] IS NOT NULL GROUP BY model, [name])) OR
        (front IS NULL AND model IN (SELECT _id FROM model WHERE front IS NULL))
      `,
        [new Date().toISOString()],
      );
    },
    note: () => {
      this.db.query(
        /* sql */ `
      UPDATE note
      SET deletedAt = ?
      WHERE
        _id IS NULL OR [key] IS NULL OR NOT json_valid([data]) OR
        model NOT IN (SELECT _id FROM model) OR
        ROWID NOT IN (SELECT ROWID FROM note GROUP BY _id, [key]) OR
        ROWID NOT IN (SELECT ROWID FROM note GROUP BY model)
      `,
        [new Date().toISOString()],
      );
    },
    card: () => {
      this.db.query(
        /* sql */ `
      UPDATE [card]
      SET deletedAt = ?
      WHERE
        (template IS NOT NULL AND note IS NULL) OR
        (template IS NULL AND note IS NOT NULL) OR
        (template IS NOT NULL AND note IS NOT NULL AND ROWID NOT IN (SELECT ROWID FROM [card] GROUP BY template, note)) OR
        (template IS NULL AND note IS NULL AND front IS NULL)
      `,
        [new Date().toISOString()],
      );
    },
  };
}
