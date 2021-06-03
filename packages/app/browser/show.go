package browser

import (
	"context"
	"errors"

	"github.com/chromedp/chromedp"
)

type AppSize chromedp.ExecAllocatorOption

func IsMaximized() AppSize {
	return chromedp.Flag("start-maximized", true)
}

// WindowSize is the command line option to set the initial window size.
func WindowSize(width, height int) AppSize {
	return chromedp.WindowSize(width, height)
}

func (b Browser) AppMode(url string, size AppSize) {
	opts := []chromedp.ExecAllocatorOption{
		chromedp.ExecPath(""),
		chromedp.NoFirstRun,
		chromedp.NoDefaultBrowserCheck,
		chromedp.Flag("app", url),
		size,
	}
	execPath := b.GetExecPath()
	if execPath != "" {
		opts = append(opts, chromedp.ExecPath(execPath))
	}

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	if e := chromedp.Run(
		ctx,
	); e != nil && !errors.Is(e, context.Canceled) {
		panic(e)
	}
}
