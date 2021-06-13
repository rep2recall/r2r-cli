# rep2recall-min

![](/docs/2021-06-13_14-23.png)

Repeat until recall, with widening intervals, minimal version.

CLI and programmability-focused memorizing flashcard app.

```
$ rep2recall --help

Repeat Until Recall - a simple, yet powerful, flashcard app

Usage:
   rep2recall {flags}
   rep2recall <command> {flags}

Commands: 
   clean                         clean the to-be-delete part of the database and exit
   help                          displays usage informationn
   load                          load the YAML into the database and exit
   proxy                         start as proxy server, for development
   quiz                          open the quiz window only
   version                       displays version number

Flags: 
   --browser                     browser to open (default: Chrome with Edge fallback) (default: .)
   --debug                       whether to run in debug mode (default: false)
   -h, --help                    displays usage information of the application or a command (default: false)
   -p, --port                    port to run the server (default: 25459)
   --server                      run in server mode (don't open the browser) (default: false)
   -v, --version                 displays version number (default: false)
```

## Simple, and file-based

You can see example input in `/data/*.yaml`. You can see that it is [Eta](https://eta.js.org/) / browser-side JavaScript based. This is further enhanced by plugins in `/packages/app/plugins/*.js`. Golang-based plugins are also possible.

Otherwise, quizzing (and mnemonic) data are generated and stored in `data.db`; with is a SQLite file. The schema can be seen in `/packages/app/db/*.go`.

## Real and latest browser-side JavaScript

You use any JavaScript that latest browsers support. Of course, `<script type="module">` is also supported.

## Better search engine

The search allows not only searching by tags (`tag:`) and data fields (`"key":`), but also by statistics (`srsLevel:0`, `wrongStreak<2`) and by date (`nextReview<-1h`).

Further design of the search engine can be seen in <https://github.com/patarapolw/qsearch>.

## Dependencies

This app utilizes Chrome DevTools Protocol. Most Windows computers already have this by default via Microsoft Edge.

However, in macOS and Linux, you will require to install either Google Chrome, or Chromium (or Ungoogled Chromium).

## Deployment as a server

You can do that, but an environment variable, `SECRET` will be required, which will be generated in `.env.local` by default.

## Parent project, as a concept

The idea is not new. It came from another project of mine, <https://github.com/rep2recall/rep2recall>.
