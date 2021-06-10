import Eta from 'eta'

import { initAPI, api } from './assets/api'

initAPI().then(async ({ ok }) => {
  if (ok) {
    const query = new URL(location.href).searchParams

    return api.get('/api/card', {
      params: {
        id: query.get('id'),
        side: query.get('side')
      }
    })
      .then(({ data: r }) => {
        return Eta.renderAsync(r.raw, r.data).catch((e) => `<pre style="background-color: red">${e}\n${JSON.stringify(r, null, 2)}</pre>`)
      })
      .then((r) => {
        document.querySelector('#Card').innerHTML = r

        document.querySelectorAll('#Card script').forEach((script) => {
          const s = document.createElement('script')
          s.innerHTML = script.innerHTML
          script.replaceWith(s)
        })
      })
  }

  throw new Error('cannot login')
}).catch((e) => {
  document.querySelector('#Card').innerHTML = `<pre style="background-color: red">${e}</pre>`
  throw e
})
