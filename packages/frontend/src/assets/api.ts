import axios from 'axios'

export const api = axios.create()

api.interceptors.response.use(undefined, async (r) => {
  if (r.response.status >= 400 && r.response.status < 500) {
    location.href = '/'
  }

  throw r
})

export async function initAPI() {
  const u = new URL(location.href)
  const token = u.searchParams.get('token')
  if (token) {
    api.defaults.headers = api.defaults.headers || {}
    api.defaults.headers['Authorization'] = `Bearer ${token}`
  }

  const { data } = await api.post<{
    ok?: boolean
  }>('/api/ok')
  return data
}
