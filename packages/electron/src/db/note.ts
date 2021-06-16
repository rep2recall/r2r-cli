import {
  Collection,
  DateType,
  Entity,
  Index,
  JsonType,
  ManyToOne,
  OneToMany,
  PrimaryKey,
  Property,
  Unique
} from '@mikro-orm/core'
import shortUUID from 'short-uuid'

import { Model } from './model'

@Entity()
export class Note {
  @PrimaryKey()
  id: string = shortUUID.generate()

  @Property({
    type: DateType
  })
  createdAt: Date = new Date()

  @Property({
    type: DateType,
    onUpdate: () => new Date()
  })
  updatedAt: Date = new Date()

  @ManyToOne(() => Model)
  model!: Model

  @OneToMany(() => NoteAttr, (attr) => attr.note)
  attrs = new Collection<NoteAttr>(this)
}

@Entity({ tableName: 'note_attr' })
@Unique({ properties: ['note', 'key'] })
export class NoteAttr {
  @PrimaryKey()
  id: string = shortUUID.generate()

  @ManyToOne(() => Note, { fieldName: 'note_id' })
  note!: Note

  @Property()
  key!: string

  @Property({
    type: JsonType
  })
  @Index({ type: 'text' })
  data!: unknown

  constructor(na: Omit<NoteAttr, 'id'>) {
    Object.assign(this, na)
  }
}
