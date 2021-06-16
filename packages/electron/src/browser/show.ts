import { BrowserWindow } from 'electron'

export async function appMode(
  url: string,
  opts: {
    width?: number
    height?: number
  } = {}
) {
  const { width, height } = opts

  const win = new BrowserWindow({
    width,
    height
  })

  if (!width || !height) {
    win.maximize()
  }

  await win.loadURL(url)
}
