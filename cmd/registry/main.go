package main

import (
	"github.com/AlbertJoey/goimx/model"
	"github.com/AlbertJoey/goimx/registry"
)

func main() {
	r := registry.New(model.NewConf())
	r.Run()
}
