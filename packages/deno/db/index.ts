import { eta, pathResolve, Sqlite, yamlParse, z } from "../deps.ts";

const zData = z.object({
  model: z.array(z.object({
    _id: z.string(),
    name: z.string().optional(),
    front: z.string().optional(),
    back: z.string().optional(),
    shared: z.string().optional(),
    generated: z.object({
      _: z.string().optional(),
    }).nonstrict(),
  })).optional(),
  template: z.array(z.object({
    _id: z.string(),
    model: z.string().optional(),
    name: z.string().optional(),
    front: z.string().optional(),
    back: z.string().optional(),
    shared: z.string().optional(),
  })).optional(),
  note: z.array(z.object({
    _id: z.string(),
    model: z.string().optional(),
    data: z.object({}).nonstrict().optional(),
  })).optional(),
  card: z.array(z.object({
    _id: z.string(),
    template: z.string().optional(),
    note: z.string().optional(),
    front: z.string().optional(),
    back: z.string().optional(),
    shared: z.string().optional(),
    mnemonic: z.string().optional(),
  })).optional(),
});

export class LocalDB {
  db: Sqlite;
  isClosed = false;

  constructor(public path = pathResolve("data.db")) {
    this.db = new Sqlite(path);

    /**
     * model
     */
    this.db.query(/* sql */ `
    CREATE TABLE IF NOT EXISTS model (
      _id           TEXT NOT NULL PRIMARY KEY,
      updatedAt     TIMESTAMP,
      deletedAt     TIMESTAMP,
      [name]        TEXT,
      front         TEXT,
      back          TEXT,
      shared        TEXT,
      [generated]   JSON
    );

    CREATE INDEX IF NOT EXISTS idx_model_updatedAt ON model (updatedAt);
    CREATE INDEX IF NOT EXISTS idx_model_deletedAt ON model (deletedAt);
    CREATE INDEX IF NOT EXISTS idx_model_name ON model ([name]);
    `);

    /**
    * template
    */
    this.db.query(/* sql */ `
    CREATE TABLE IF NOT EXISTS template (
      _id           TEXT NOT NULL PRIMARY KEY,
      updatedAt     TIMESTAMP,
      deletedAt     TIMESTAMP,
      model         TEXT NOT NULL REFERENCES model(_id) ON DELETE CASCADE,
      [name]        TEXT,  -- should not duplicate, although not restricted; otherwise, name can't be null
      front         TEXT,
      back          TEXT,
      shared        TEXT
    );

    CREATE INDEX IF NOT EXISTS idx_template_updatedAt ON template (updatedAt);
    CREATE INDEX IF NOT EXISTS idx_template_deletedAt ON template (deletedAt);
    CREATE INDEX IF NOT EXISTS idx_template_model ON template (model);
    CREATE INDEX IF NOT EXISTS idx_template_name ON template ([name]);
    `);

    /**
    * note
    */
    this.db.query(/* sql */ `
    CREATE TABLE IF NOT EXISTS note USING fts5(
      _id,          -- TEXT NOT NULL
      updatedAt,    -- TIMESTAMP
      deletedAt,    -- TIMESTAMP
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
      updatedAt     TIMESTAMP,
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

    CREATE INDEX IF NOT EXISTS idx_card_updatedAt ON [card] (updatedAt);
    CREATE INDEX IF NOT EXISTS idx_card_deletedAt ON [card] (deletedAt);
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

  clean() {
    for (const k of Object.keys(this.tidy)) {
      this.tidy[k as keyof LocalDB["tidy"]]();

      this.db.query(
        /* sql */ `
      DELETE FROM [${k}]
      WHERE deletedAt <= ?
      `,
        [new Date().toISOString()],
      );
    }
  }

  tidy = {
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

  async doLoad(filename: string) {
    const updatedAt = new Date().toISOString();

    const dataFile = await zData.parseAsync(yamlParse(
      await Deno.readTextFile(pathResolve(filename)),
    ));

    const modelGenMap = new Map<string, Record<string, unknown>>();

    if (dataFile.model) {
      dataFile.model.map((m) => {
        if (m.generated) {
          modelGenMap.set(m._id, m.generated);
        }

        this.db.query(
          /* sql */ `
        INSERT OR REPLACE INTO model (_id, [name], front, back, shared, [generated], updatedAt)
        VALUES (:_id, :name, :front, :back, :shared, :generated, :updatedAt)
        `,
          {
            _id: m._id,
            name: m.name,
            front: m.front,
            back: m.back,
            shared: m.shared,
            generated: m.generated ? JSON.stringify(m.generated) : null,
            updatedAt,
          },
        );
      });
    }

    if (dataFile.template) {
      dataFile.template.map((t) => {
        this.db.query(
          /* sql */ `
        INSERT OR REPLACE INTO template (_id, model, [name], front, back, shared, updatedAt)
        VALUES (:_id, :model, :name, :front, :back, :shared, :updatedAt)
        `,
          {
            _id: t._id,
            model: t.model,
            name: t.name,
            front: t.front,
            back: t.back,
            shared: t.shared,
            updatedAt,
          },
        );
      });
    }

    if (dataFile.note) {
      await Promise.all(dataFile.note.map(async (n) => {
        let gen: Record<string, unknown> | null = null;

        if (n.model) {
          gen = modelGenMap.get(n.model) || null;

          if (!gen) {
            const [r] = [
              ...this.db.query(
                /* sql */ `
            SELECT [generated] FROM model WHERE _id = ?
            `,
                [n.model],
              ),
            ].map(([generated]) => generated);

            if (r) {
              gen = JSON.parse(r);
              modelGenMap.set(n.model, gen || {});
            }
          }
        }

        const data = n.data || {};

        if (gen) {
          const { _, gen0 } = gen;

          await Promise.all(
            Object.entries(gen0 as Record<string, unknown>).map(
              async ([k, v]) => {
                if (typeof v === "string") {
                  (gen0 as Record<string, unknown>)[k] = await eta.renderAsync(
                    v,
                    data,
                  );
                } else {
                  (gen0 as Record<string, unknown>)[k] = v;
                }
              },
            ),
          );

          Object.assign(data, {
            ...(gen0 as Record<string, unknown>),
            ...data,
          });

          if (typeof _ === "string") {
            await eta.renderAsync(_, data);
          }
        }

        Object.entries(data).map(([key, v]) => {
          if (!v) {
            return;
          }

          this.db.query(
            /* sql */ `
          UPDATE note SET deletedAt = :updatedAt
          WHERE _id = :_id AND model = :model AND [key] = :key
          `,
            {
              _id: n._id,
              model: n.model,
              key,
              updatedAt,
            },
          );

          this.db.query(
            /* sql */ `
          INSERT INTO note (_id, model, [key], [data], [generated], updatedAt)
          VALUES (:_id, :model, :key, :data, :generated, :updatedAt)
          `,
            {
              _id: n._id,
              model: n.model,
              key,
              data: JSON.stringify(v),
              generated: !!((n.data || {}) as Record<string, unknown>)[key],
              updatedAt,
            },
          );
        });
      }));

      this.db.query(
        /* sql */ `
      DELETE FROM note WHERE deletedAt = :updatedAt
      `,
        { updatedAt },
      );
    }

    if (dataFile.card) {
      dataFile.card.map((c) => {
        /**
         * DO NOT prompty delete review records
         */
        this.db.query(
          /* sql */ `
        UPDATE [card] SET deletedAt = :updatedAt, _id = '_'||_id
        WHERE _id = :_id
        `,
          {
            _id: c._id,
            updatedAt,
          },
        );

        this.db.query(
          /* sql */ `
        INSERT INTO [card] (_id, template, note, front, back, shared, mnemonic, updatedAt)
        VALUES (:_id, :template, :note, :front, :back, :shared, :mnemonic, :updatedAt)
        ON CONFLICT DO UPDATE SET
          template = :template,
          note = :note,
          front = :front,
          back = :back,
          shared = :shared,
          mnemonic = :mnemonic,
          updatedAt = :updatedAt
        `,
          {
            _id: c._id,
            template: c.template,
            note: c.note,
            front: c.front,
            back: c.back,
            shared: c.shared,
            mnemonic: c.mnemonic,
            updatedAt,
          },
        );
      });
    }
  }
}
