import {
  DateType,
  Entity,
  Index,
  JsonType,
  PrimaryKey,
  Property
} from '@mikro-orm/core'
import shortUUID from 'short-uuid'

@Entity()
export class Model {
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

  @Property()
  @Index()
  name: string = ''

  @Property()
  front: string = ''

  @Property()
  back: string = ''

  @Property()
  shared: string = ''

  @Property({
    type: JsonType
  })
  generated: {
    _?: string
  } & Record<string, unknown> = {}
}
