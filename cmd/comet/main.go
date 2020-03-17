package main

import (
	"github.com/AlbertJoey/goimx/comet"
	"github.com/AlbertJoey/goimx/model"
)

func main() {
	c := comet.New(model.NewConf())
	c.Run()
}
