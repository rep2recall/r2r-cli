import fs from 'fs'
import path from 'path'
import qs from 'querystring'
import { PassThrough } from 'stream'

import { MikroORM } from '@mikro-orm/core'
import { app as electron } from 'electron'
import fastify, { FastifyInstance } from 'fastify'
import cors from 'fastify-cors'
import fastifyStatic from 'fastify-static'
import pino from 'pino'
import stripANSIStream from 'string-ansi-stream'

import { initDatabase } from './db'
import { ROOTDIR, g } from './shared'

export interface ServerOptions {
  isServer: boolean
  debug: boolean
  proxy: boolean
  port: number
}

interface ServerInstance {
  app: FastifyInstance
  logger: pino.Logger
  orm: MikroORM
}

export class Server implements ServerInstance {
  static async init(opts: ServerOptions): Promise<Server> {
    const logThrough = new PassThrough()
    const logger = pino(
      {
        prettyPrint: opts.proxy,
        serializers: {
          req(req) {
            const [url, q] = req.url.split(/\?(.+)$/)
            const query = q ? qs.parse(q) : undefined

            return { method: req.method, url, query, hostname: req.hostname }
          }
        }
      },
      logThrough
    )

    logThrough
      .pipe(stripANSIStream())
      .pipe(fs.createWriteStream(path.resolve(ROOTDIR, 'server.log')))
    logThrough.pipe(process.stdout)

    const app = fastify({
      logger
    })

    app.addHook('preHandler', async (req) => {
      if (req.body) {
        req.log.info({ body: req.body }, 'parsed body')
      }

      return null
    })

    if (opts.proxy) {
      app.register(cors)
    }

    app.register(fastifyStatic, {
      root: path.resolve(ROOTDIR, 'public')
    })

    await new Promise<void>((resolve, reject) => {
      app.listen(opts.port, (err) => {
        if (err) {
          reject(err)
          return
        }

        resolve()
      })
    })

    g.server = new this({
      app,
      logger,
      orm: await initDatabase(
        path.join(electron.getPath('userData'), 'data.db')
      )
    })

    return g.server
  }

  app: FastifyInstance
  logger: pino.Logger
  orm: MikroORM

  private isClosed = false

  private constructor(it: ServerInstance) {
    this.app = it.app
    this.logger = it.logger
    this.orm = it.orm
  }

  async close() {
    if (!this.isClosed) {
      this.isClosed = true
      await this.orm.close()
      await this.app.close()
    }
  }
}
