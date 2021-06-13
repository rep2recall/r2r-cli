import {
  Collection,
  DateType,
  Entity,
  Index,
  JsonType,
  ManyToOne,
  OneToMany,
  PrimaryKey,
  Property
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
export class NoteAttr {
  @PrimaryKey()
  id!: number

  @Property({
    type: DateType
  })
  createdAt: Date = new Date()

  @Property({
    type: DateType,
    onUpdate: () => new Date()
  })
  updatedAt: Date = new Date()

  @ManyToOne(() => Note)
  @Index()
  note!: Note

  @Property()
  @Index()
  key!: string

  @Property({
    type: JsonType
  })
  @Index({ type: 'text' })
  data!: unknown
}
