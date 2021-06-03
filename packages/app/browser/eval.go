package browser

import (
	"context"
	"encoding/json"
	"fmt"
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
	Port    int
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

	pluginb, e := json.Marshal(strings.Join(opts.Plugins, ";\n"))
	if e != nil {
		panic(e)
	}

	actions := []chromedp.Action{
		chromedp.Navigate(fmt.Sprintf("http://localhost:%d/script.html", opts.Port)),
		chromedp.WaitReady("body"),
		chromedp.EvaluateAsDevTools(fmt.Sprintf(`
		s = document.createElement('script');
		s.type = "module";
		s.innerHTML = %s;
		document.body.append(s);`, string(pluginb)), nil),
		chromedp.WaitReady("body"),
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
			el.innerHTML += '<br/>';
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
