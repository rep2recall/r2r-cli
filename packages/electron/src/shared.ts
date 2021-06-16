import fs from 'fs'
import path from 'path'

import { app as electron } from 'electron'
import yaml from 'js-yaml'
import S from 'jsonschema-definer'

import type { Server } from './server'

export const ROOTDIR = path.resolve(path.dirname(__dirname))

export const g = new (class {
  server!: Server
  config!: typeof sConfig.type
})()

const sConfig = S.shape({
  port: S.integer().maximum(65536).minimum(1),
  db: S.string().minLength(1)
})

export function config() {
  const USER_DATA_PATH = electron.getPath('userData')
  const ASSETS_PATH = path.join(ROOTDIR, 'assets')
  const FILENAME = 'config.yaml'

  if (!fs.existsSync(path.join(USER_DATA_PATH, FILENAME))) {
    fs.copyFileSync(path.join(ASSETS_PATH, FILENAME), path.join(USER_DATA_PATH, FILENAME))
  }

  g.config = sConfig.ensure(yaml.load(fs.readFileSync(path.join(USER_DATA_PATH, FILENAME), 'utf-8')) as any)

  return g.config
}
