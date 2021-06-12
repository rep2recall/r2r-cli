import { createApp } from 'vue'
import Init from './assets/Init.vue'
import { api, initAPI } from './assets/api'

async function main() {
  const { ok } = await initAPI()
  if (!ok) {
    return
  }

  const { data } = await api.post('/api/quiz/init', undefined, {
    params: {
      q: new URL(location.href).searchParams.get('q'),
      state: 'new,learning,due',
    },
  })

  createApp(Init, {
    type: 'Quiz',
    session: data.id,
    close: true
  }).mount('#Quiz')
}

main()
