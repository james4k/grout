package main

import (
	"github.com/james4k/grout"
	_ "github.com/james4k/grout/listing"
	"runtime"
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU() * 2)
	grout.Build("", "",
		&grout.Options{
			Verbose:  true,
			HttpHost: ":8000",
		})
}
