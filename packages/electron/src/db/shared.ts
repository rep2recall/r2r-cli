import { Platform, Type } from '@mikro-orm/core'
import { SqlitePlatform } from '@mikro-orm/sqlite'

export class LikeableArrayType extends Type<string[], string> {
  constructor(private separator = ' ') {
    super()
  }

  convertToDatabaseValue(value: string[] | string, platform: Platform): string {
    if (platform instanceof SqlitePlatform) {
      if (Array.isArray(value)) {
        if (value.length) {
          return this.separator + value.join(this.separator) + this.separator
        }

        return ''
      }
    }

    return value as string
  }

  convertToJSValue(value: string[] | string, platform: Platform): string[] {
    if (platform instanceof SqlitePlatform) {
      if (typeof value === 'string') {
        if (value[0] === this.separator) {
          value = value.slice(1)
        }
        if (value[value.length - 1] === this.separator) {
          value = value.slice(0, value.length - 1)
        }

        return value.split(this.separator)
      }
    }

    return (value as string[]) || []
  }

  getColumnType() {
    return 'TEXT'
  }
}
