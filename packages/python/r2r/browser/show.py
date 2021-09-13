import pyppeteer


async def app_mode(b: str, url: str, width=800, height=600, is_maximized=True) -> None:
    args = [f"app={url}"]
    if is_maximized:
        args += ["start-maximized=true"]
    else:
        args += [f"window-size={width},{height}"]

    browser = await pyppeteer.launch(executablePath=b, headless=False, args=args)
    await browser.close()
