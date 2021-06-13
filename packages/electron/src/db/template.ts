import {
  DateType,
  Entity,
  Index,
  ManyToOne,
  PrimaryKey,
  Property
} from '@mikro-orm/core'
import shortUUID from 'short-uuid'

import { Model } from './model'

@Entity()
export class Template {
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

  @ManyToOne(() => Model, { nullable: true })
  @Index()
  model?: Model

  @Property()
  @Index()
  name: string = ''

  @Property()
  front: string = ''

  @Property()
  back: string = ''

  @Property()
  shared: string = ''

  @Property()
  if: string = ''
}
