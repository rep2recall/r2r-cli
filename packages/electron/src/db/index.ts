import { MikroORM } from '@mikro-orm/core'

import { Card } from './card'
import { Model } from './model'
import { Note, NoteAttr } from './note'
import { Template } from './template'

export async function initDatabase(filename: string) {
  const MONGO_URI = process.env.MONGO_URI

  const orm = await MikroORM.init({
    entities: [Model, Template, Note, NoteAttr, Card],
    type: MONGO_URI ? 'mongo' : 'sqlite',
    dbName: MONGO_URI ? 'rep2recall' : filename,
    clientUrl: MONGO_URI,
    implicitTransactions: true,
    ensureIndexes: true
  })

  if (!MONGO_URI) {
    await orm.em.getConnection().execute(/* sql */ `
    CREATE VIRTUAL TABLE IF NOT EXISTS note_fts USING fts5(
      note_id UNINDEXED,
      "key",
      "data",
      content=note_attr,
      content_rowid=id,
      tokenize=porter
    );

    -- Triggers to keep the FTS index up to date.
    CREATE TRIGGER IF NOT EXISTS t_note_attr_ai AFTER INSERT ON note_attr BEGIN
      INSERT INTO note_fts(rowid, note_id, key, data) VALUES (new.rowid, new.note_id, new.key, new.data);
    END;
    CREATE TRIGGER IF NOT EXISTS t_note_attr_ad AFTER DELETE ON note_attr BEGIN
      INSERT INTO note_fts(note_fts, rowid, note_id, key, data) VALUES ('delete', old.rowid, old.note_id, old.key, old.data);
    END;
    CREATE TRIGGER IF NOT EXISTS t_note_attr_au AFTER UPDATE ON note_attr BEGIN
      INSERT INTO note_fts(note_fts, rowid, note_id, key, data) VALUES ('delete', old.rowid, old.note_id, old.key, old.data);
      INSERT INTO note_fts(rowid, note_id, key, data) VALUES (new.rowid, new.note_id, new.key, new.data);
    END;
    `)
  } else {
    await (orm.em.getDriver() as any)?.createCollections?.()
  }

  return orm
}
