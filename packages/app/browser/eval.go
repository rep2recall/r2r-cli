package browser

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/chromedp/chromedp"
)

type EvalContext struct {
	JS     string
	Output interface{}
}

func (b Browser) Eval(imports []string, scripts ...*EvalContext) {
	opts := chromedp.DefaultExecAllocatorOptions[:]
	execPath := b.GetExecPath()
	if execPath != "" {
		opts = append(opts, chromedp.ExecPath(execPath))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	for i, im := range imports {
		imports[i] = strings.ReplaceAll(im, "\"", "\\\"")
	}

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
			<script type="module">
			import "%s";
			</script>
		</body>
		</html>
		`, strings.Join(imports, "\";\nimport \"")))),
	}
	for _, s := range scripts {
		actions = append(actions, chromedp.Evaluate(s.JS, &s.Output))
	}

	if e := chromedp.Run(
		ctx,
		actions...,
	); e != nil {
		panic(e)
	}
}
