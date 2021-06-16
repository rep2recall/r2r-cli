```
rep2recall [cmd] [args]

Commands:
  rep2recall [files...]        open in GUI mode, for full interaction  [default]
  rep2recall load <files...>   load files into the database
  rep2recall clean [files...]  clean the to-be-deleted part of the database and
                               exit
  rep2recall server            open in server mode, for online deployment

Positionals:
  files                                                                 [string]

Options:
      --version  Show version number                                   [boolean]
  -p, --port     port to run the server                [number] [default: 25459]
      --db       path to the database, or MONGO_URI
                                                 [string] [default: "./data.db"]
      --debug    whether to run in debug mode                          [boolean]
      --help     Show help                                             [boolean]
  -f, --filter   keyword to filter                                      [string]
```

Furthermore, environmental variable `USER_DATA_DIR` and `$USER_DATA_DIR/config.yaml` can be tweaked.
