import {
  DateType,
  Entity,
  Index,
  JsonType,
  ManyToOne,
  PrimaryKey,
  Property,
  Unique
} from '@mikro-orm/core'
import shortUUID from 'short-uuid'

import { Note } from './note'
import { LikeableArrayType } from './shared'
import { Template } from './template'

@Entity()
@Unique({ properties: ['template', 'note'] })
export class Card {
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

  @ManyToOne(() => Template, { nullable: true })
  @Index()
  template?: Template

  @ManyToOne(() => Note, { nullable: true })
  @Index()
  note?: Note

  @Property()
  front: string = ''

  @Property()
  back: string = ''

  @Property()
  shared: string = ''

  @Property({
    type: JsonType
  })
  mnemonic: unknown = {}

  @Property()
  @Index()
  srsLevel: number = 0

  @Property({
    type: DateType,
    nullable: true
  })
  @Index()
  nextReview?: Date

  @Property({
    type: DateType,
    nullable: true
  })
  @Index()
  lastRight?: Date

  @Property({
    type: DateType,
    nullable: true
  })
  @Index()
  lastWrong?: Date

  @Property()
  @Index()
  maxRight: number = 0

  @Property()
  @Index()
  maxWrong: number = 0

  @Property()
  @Index()
  rightStreak: number = 0

  @Property()
  @Index()
  wrongStreak: number = 0

  @Property({
    type: LikeableArrayType
  })
  @Index()
  tag: string[] = []
}
