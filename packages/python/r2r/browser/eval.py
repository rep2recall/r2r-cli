from dataclasses import dataclass
import json
from typing import Any

import pyppeteer


@dataclass
class EvalContext:
    js: str
    output: Any = None


async def eval(
    browser,
    port: int,
    scripts: list[EvalContext],
    plugins: list[str] = [],
):
    is_new_browser = False
    if browser is None:
        browser = await pyppeteer.launch()
        is_new_browser = True

    page = await browser.newPage()
    await page.goto(f"http://localhost:{port}/script")
    await page.evaluate(
        f"""
            s = document.createElement('script');
            s.type = "module";
            s.innerHTML = {json.dumps("\n".join(plugins))};
            document.body.append(s);
        """
    )

    await page.once("load")

    for i, s in enumerate(scripts):
        page.evaluate(
            f"""
            (async () => {{
                const r = {s.js};
                return r;
            }})().then(r => {{
                __output['{i}'] = r;
                document.querySelector('#output').innerText = JSON.stringify(__output, null, 2);
                if (Object.keys(__output).length === {len(scripts)}) document.querySelector('#output').setAttribute('selected', '')
            }}).catch(e => {{
                const el = document.querySelector('#error');
                el.innerText += e;
                el.innerHTML += '<br/>';
                el.style.display = 'block';
            }})
            """
        )

    await page.waitForSelector("#output[selected]")

    for i, s in enumerate(scripts):
        s.output = await page.evaluate(f"__output['{i}']")

    await page.close()

    if is_new_browser:
        await browser.close()
