package main

import (
	"github.com/CaLYpSEDamir/InfoTechGo/loader"
)

func main() {
	loading := loader.Loader{}
	loading.Init()
	loading.Load()
}
