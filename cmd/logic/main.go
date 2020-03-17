package main

import (
	"github.com/AlbertJoey/goimx/logic"
	"github.com/AlbertJoey/goimx/model"
)

func main() {
	l := logic.New(model.NewConf())
	l.Run()
}
