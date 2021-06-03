package browser

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/chromedp/chromedp"
)

// EvalContext - Container for eval context, and capture the result
//
// @see Eval
type EvalContext struct {
	JS     string
	Output interface{}
}

type EvalOptions struct {
	Plugins []string
	Visible bool
}

// Eval - Evaluate JavaScript in EvalContext
func (b Browser) Eval(scripts []*EvalContext, opts EvalOptions) {
	args := chromedp.DefaultExecAllocatorOptions[:]
	if opts.Visible {
		newArgs := make([]chromedp.ExecAllocatorOption, 0)
		for i, a := range args {
			if i != 2 {
				newArgs = append(newArgs, a)
			}
		}
		args = newArgs
	}

	execPath := b.GetExecPath()
	if execPath != "" {
		args = append(args, chromedp.ExecPath(execPath))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), args...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	actions := []chromedp.Action{
		chromedp.Navigate("data:text/html," + url.PathEscape(fmt.Sprintf(`
		<!DOCTYPE html>
		<html lang="en">
		<head>
			<base href=".">
			<meta charset="UTF-8">
			<meta http-equiv="X-UA-Compatible" content="IE=edge">
		</head>
		<body>
			<div id="error" style="display: none; background-color: red"></div>
			<pre id="output"></pre>
			<script type="module">
			%s;
			window.__output = {};
			</script>
		</body>
		</html>
		`, strings.Join(opts.Plugins, "\n")))),
	}
	for i, s := range scripts {
		js := fmt.Sprintf(`(async () => {
			const r = %s;
			return r;
		})().then(r => {
			__output['%d'] = r;
			document.querySelector('#output').innerText = JSON.stringify(__output, null, 2);
			if (Object.keys(__output).length === %d) document.querySelector('#output').setAttribute('selected', '')
		}).catch(e => {
			const el = document.querySelector('#error');
			el.innerText += e;
			el.style.display = 'block';
		})`, s.JS, i, len(scripts))

		actions = append(actions, chromedp.Evaluate(js, &s.Output))
	}
	actions = append(actions, chromedp.WaitSelected("#output"))
	for i, s := range scripts {
		actions = append(actions, chromedp.Evaluate(fmt.Sprintf("__output['%d']", i), &s.Output))
	}

	if e := chromedp.Run(
		ctx,
		actions...,
	); e != nil {
		panic(e)
	}
}
