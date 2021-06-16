import path from 'path'
import qs from 'querystring'

import { app as electron } from 'electron'
import yargs from 'yargs'

import { appMode } from './browser/show'
import { Server } from './server'
import { config, g } from './shared'
import { load } from './util/load'

function setAppDir() {
  const APP_NAME = 'rep2recall'
  const { USER_DATA_DIR } = process.env

  if (USER_DATA_DIR) {
    electron.setPath('userData', USER_DATA_DIR)
  } else {
    electron.setPath('userData', path.join(electron.getPath('appData'), APP_NAME))
  }
}

async function main() {
  setAppDir()
  await electron.whenReady()
  setAppDir()

  const cfg = config()

  const args: {
    cmd?: 'load' | 'quiz' | 'clean' | 'server'
    files?: string[]
    filter?: string
    proxy?: boolean
  } = {}

  const { argv } = yargs
    .scriptName('rep2recall')
    .usage('$0 [cmd] [args]')
    .command(
      '$0 [files...]',
      'open in GUI mode, for full interaction',
      (yargs) => {
        return yargs
          .positional('files', {
            type: 'string',
            demandOption: false,
            normalize: true,
            array: true
          })
          .option('filter', {
            alias: 'f',
            describe: 'keyword to filter',
            type: 'string'
          })
      },
      ({ files, filter }) => {
        args.files = files
        args.filter = filter
      }
    )
    .command(
      'load <files...>',
      'load files into the database',
      (yargs) => {
        return yargs.positional('files', {
          type: 'string',
          demandOption: true,
          normalize: true,
          array: true
        })
      },
      ({ files }) => {
        args.cmd = 'load'
        args.files = files
      }
    )
    .command(
      'clean [files...]',
      'clean the to-be-deleted part of the database and exit',
      (yargs) => {
        return yargs
          .positional('files', {
            type: 'string',
            demandOption: false,
            normalize: true,
            array: true
          })
          .option('filter', {
            alias: 'f',
            describe: 'keyword to filter',
            type: 'string'
          })
      },
      ({ files, filter }) => {
        args.cmd = 'clean'
        args.files = files
        args.filter = filter
      }
    )
    .command(
      'server',
      'open in server mode, for online deployment',
      (yargs) => {
        return yargs.option('proxy', {
          describe: 'use as proxy server (enable CORS)',
          type: 'boolean'
        })
      },
      ({ proxy }) => {
        args.cmd = 'server'
        args.proxy = proxy
      }
    )
    .option('port', {
      alias: 'p',
      describe: 'port to run the server',
      type: 'number',
      default: cfg.port
    })
    .option('db', {
      describe: 'path to the database, or MONGO_URI',
      type: 'string',
      default: cfg.db
    })
    .option('debug', {
      describe: 'whether to run in debug mode',
      type: 'boolean'
    })
    .help()

  const { cmd, files = [], filter: q = '', proxy = false } = args
  const { port, debug = false, db } = argv

  g.config.port = port
  g.config.db = db

  await Server.init({
    isServer: true,
    debug,
    proxy,
    port,
  })

  electron.on('window-all-closed', () => {
    electron.quit()
  })

  switch (cmd) {
    case undefined:
      appMode(`http://localhost:${port}/app?${qs.stringify({
        q,
        files: files.length ? JSON.stringify(files) : undefined
      })}`)
      break
    case 'quiz':
      appMode(`http://localhost:${port}/quiz?${qs.stringify({
        q,
        files: files.length ? JSON.stringify(files) : undefined,
      })}`, {
        width: 600,
        height: 800
      })
      break
    case 'load':
      for (const f of files) {
        await load(f, {
          debug,
          port
        })
      }
      electron.quit()
      break
    case 'clean':
      electron.quit()
      break
    default:
      electron.quit()
  }
}

main()
