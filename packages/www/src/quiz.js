import { initAPI } from './assets/api'

initAPI().then(() => {
    document.querySelector('#Quiz').innerText = 'Logged in'
})
