import Eta from 'eta'

import { initAPI, api } from './assets/api'

initAPI().then(async ({ ok }) => {
  if (ok) {
    const query = new URL(location.href).searchParams

    return api.get('/api/item', {
      params: {
        id: query.get('id'),
        side: query.get('side')
      }
    })
      .then(({ data: r }) => {
        console.log(r)
        return Eta.renderAsync(r.raw, r.data)
      })
      .then((r) => {
        document.querySelector('#Item').innerHTML = r
      })
  }

  throw new Error('cannot login')
}).catch((e) => {
  document.querySelector('#Item').innerHTML = `<pre style="background-color: red">${e}</pre>`
  throw e
})
