package controllers

import (
	"aahframe.work/aah"
)

// GoModController handles `go mod` requests, this is gonna be future package management way.
type GoModController struct {
	*aah.Context
}

// Handle ..
func (c *GoModController) Handle(modPath string) {
	// fmt.Println("modPath", modPath)
	// fmt.Println(strings.Index(modPath, "/@v"))
	c.Reply().Text("go mod handling coming soon...")
}
