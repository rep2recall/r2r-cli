declare global {
  interface Window {
    Eta: typeof import('eta')
  }
}

const script = document.createElement('script')
script.src = '/vendor/eta/eta.min.js'
document.body.append(script)

const Eta = window.Eta

export default Eta
