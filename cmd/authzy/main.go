package main

import (
	"github.com/sean-/seed"

	"github.com/zbiljic/authzy/cmd/authzy/internal"
)

func init() {
	seed.Init() //nolint
}

func main() {
	internal.Execute()
}
