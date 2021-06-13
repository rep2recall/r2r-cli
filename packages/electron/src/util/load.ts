import S from 'jsonschema-definer'

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
