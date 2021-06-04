import { initAPI } from './assets/api'

initAPI().then(() => {
    document.querySelector('#App').innerText = 'Logged in'
})
