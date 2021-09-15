package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/yanyiwu/gojieba"
)

func main() {
	x := gojieba.NewJieba()
	defer x.Free()
	fmt.Println(strings.Join(x.CutForSearch(os.Args[len(os.Args)-1], true), " "))
}
