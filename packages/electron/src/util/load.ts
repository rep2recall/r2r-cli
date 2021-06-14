import fs from 'fs'
import path from 'path'

import { app as electron } from 'electron'
import glob from 'fast-glob'
import yaml from 'js-yaml'
import S from 'jsonschema-definer'

import { EvalContext, evaluate } from '../browser/eval'
import { Model } from '../db/model'
import { Note, NoteAttr } from '../db/note'
import { g } from '../shared'

const sFile = S.shape({
  model: S.list(
    S.shape({
      id: S.string(),
      name: S.string().optional(),
      front: S.string().optional(),
      back: S.string().optional(),
      shared: S.string().optional(),
      generated: S.object().custom(
        (v) => typeof v._ === 'undefined' || typeof v._ === 'string'
      )
    })
  ).optional(),
  template: S.list(
    S.shape({
      id: S.string(),
      modelId: S.string(),
      name: S.string().optional(),
      front: S.string().optional(),
      back: S.string().optional(),
      shared: S.string().optional(),
      if: S.string().optional()
    })
  ).optional(),
  note: S.list(
    S.shape({
      id: S.string(),
      modelId: S.string(),
      data: S.object()
    })
  ).optional(),
  card: S.list(
    S.shape({
      id: S.string(),
      templateId: S.string().optional(),
      noteId: S.string().optional(),
      tag: S.list(S.string()).optional(),
      front: S.string().optional(),
      back: S.string().optional(),
      shared: S.string().optional()
    }).custom((v) => (v.templateId ? v.noteId : true))
  ).optional()
})

export type IFile = typeof sFile.type

export function loadFile(f: string): IFile {
  return sFile.ensure(yaml.load(fs.readFileSync(f, 'utf-8')) as any)
}

export async function load(
  f: string,
  opts: {
    debug?: boolean
    port: number
  }
) {
  const userData = electron.getPath('userData')
  const loadedData = loadFile(f)

  const modelGenMap: Record<string, Record<string, unknown>> = {}

  if (loadedData.model && loadedData.model.length) {
    const models = await g.server.orm.em
      .find(
        Model,
        {
          id: { $in: loadedData.model.map((m) => m.id) }
        },
        { fields: ['id'] }
      )
      .then((rs) =>
        rs.reduce(
          (prev, m) => ({ ...prev, [m.id]: m }),
          {} as Record<string, Model>
        )
      )

    for (const m of loadedData.model) {
      if (m.generated) {
        modelGenMap[m.id] = m.generated
      }

      let model = models[m.id]
      if (model) {
        Object.assign(
          model,
          JSON.parse(
            JSON.stringify({
              name: m.name,
              front: m.front,
              back: m.back,
              shared: m.shared,
              generated: m.generated
            })
          )
        )
      } else {
        model = new Model({
          name: m.name,
          front: m.front,
          back: m.back,
          shared: m.shared,
          generated: m.generated
        })
      }

      g.server.orm.em.persist(model)
    }

    await g.server.orm.em.flush()
  }

  const plugins = await glob(['plugins/www/*.js'], {
    cwd: userData
  }).then((ps) =>
    ps.map((p) => `import "/${encodeURI(p.replace(path.sep, '/'))}";`)
  )

  if (loadedData.note && loadedData.note.length) {
    const toGenerate: EvalContext<{
      id: string
      data: Record<string, unknown>
    }>[] = []

    await g.server.orm.em
      .find(
        Model,
        {
          $and: [
            {
              id: { $in: [...new Set(loadedData.note.map((m) => m.modelId))] }
            },
            { id: { $nin: Object.keys(modelGenMap) } }
          ]
        },
        {
          fields: ['id', 'generated']
        }
      )
      .then((rs) =>
        rs.map((m) => {
          modelGenMap[m.id] = m.generated
        })
      )

    for (const n of loadedData.note) {
      if (modelGenMap[n.modelId] && modelGenMap[n.modelId]._) {
        toGenerate.push({
          js: /* js */ `
          (async function() {
						const data = ${JSON.stringify(n.data)};
						await Eta.renderAsync(${JSON.stringify(modelGenMap[n.modelId]._)}, data);
						return {
							id: ${JSON.stringify(n.id)},
							data
						};
					})();`
        })
      }
    }

    const noteGenResultMap: Record<string, Record<string, unknown>> = {}

    if (toGenerate.length > 0) {
      await evaluate(toGenerate, {
        plugins,
        port: opts.port,
        visible: opts.debug
      })

      toGenerate.map((ctx) => {
        if (ctx.output) {
          noteGenResultMap[ctx.output.id] = ctx.output.data
        }
      })
    }

    const notes = await g.server.orm.em
      .find(
        Note,
        {
          id: { $in: loadedData.note.map((m) => m.id) }
        },
        { fields: ['id'] }
      )
      .then((rs) =>
        rs.reduce(
          (prev, m) => ({ ...prev, [m.id]: m }),
          {} as Record<string, Note>
        )
      )

    for (const n of loadedData.note) {
      let note = notes[n.id]
      if (!note) {
        note = new Note()
        note.model = g.server.orm.em.getReference(Model, n.modelId)
      }

      const attrs: NoteAttr[] = Object.entries(
        noteGenResultMap[n.id] || n.data
      ).map(
        ([k, v]) =>
          new NoteAttr({
            note: g.server.orm.em.getReference(Note, n.id),
            key: k,
            data: v
          })
      )

      note.attrs.removeAll()
      note.attrs.add(...attrs)

      g.server.orm.em.persist(note)
    }

    await g.server.orm.em.flush()
  }
}
