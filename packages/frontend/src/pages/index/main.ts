import axios from 'axios'

const form = document.querySelector('form') as HTMLFormElement
const input = document.querySelector('input[name="secret"]') as HTMLInputElement

form.onsubmit = (e) => {
  e.preventDefault()

  if (!input.value) {
    return
  }

  axios.post<{
    token: string
  }>('/server/login', undefined, {
    auth: {
      username: 'DEFAULT',
      password: input.value
    }
  }).then(({ data: { token } }) => {
    location.href = '/app?token=' + token
  })
}