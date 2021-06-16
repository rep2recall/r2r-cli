import ON_DEATH from 'death'
import yargs from 'yargs'

import { Server } from './server'

const { argv } = yargs
  .scriptName('rep2recall')
  .usage('$0 [cmd] [args]')
  .command(
    '$0 [files...]',
    'open in GUI mode, for full interaction',
    (yargs) => {
      yargs
        .positional('files', {
          type: 'string',
          demandOption: false,
          normalize: true
        })
        .option('filter', {
          alias: 'f',
          describe: 'keyword to filter',
          type: 'string'
        })
    },
    (argv) => {
      console.log(argv)
    }
  )
  .command(
    'load <files...>',
    'load files into the database',
    (yargs) => {
      yargs.positional('files', {
        type: 'string',
        demandOption: true,
        normalize: true
      })
    },
    (argv) => {
      console.log(argv)
    }
  )
  .command(
    'clean [files...]',
    'clean the to-be-deleted part of the database and exit',
    (yargs) => {
      yargs
        .positional('files', {
          type: 'string',
          demandOption: false,
          normalize: true
        })
        .option('filter', {
          alias: 'f',
          describe: 'keyword to filter',
          type: 'string'
        })
    },
    (argv) => {
      console.log(argv)
    }
  )
  .command(
    'server',
    'open in Server mode, for online deployment',
    (yargs) => {
      yargs.option('proxy', {
        describe: 'use as proxy server (enable CORS)',
        type: 'boolean'
      })
    },
    (argv) => {
      console.log(argv)
    }
  )
  .option('port', {
    alias: 'p',
    describe: 'port to run the server',
    type: 'number',
    default: 25459
  })
  .option('debug', {
    describe: 'whether to run in debug mode',
    type: 'boolean'
  })
  .help()

async function main() {
  console.log(argv)

  if (argv._[0] === 'server') {
    const srv = await Server.init({
      isServer: true,
      debug: false,
      proxy: false,
      port: argv.port as number,
    })

    ON_DEATH(() => {
      srv.close()
    })

    return
  }

  process.exit(0)
}

main()
