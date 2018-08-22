package controllers

import (
	"gorepositree.com/app/proxypass"
	"gorepositree.com/app/vanity"

	"aahframe.work/aah"
)

// VanityController handles the classic `go get` handling, gonna become legacy.
type VanityController struct {
	*aah.Context
}

// Handle method handles Go vanity package request. If not found then it passes
// control over to proxy pass.
func (c *VanityController) Handle(vanityPath string) {
	pkg := vanity.Lookup(c.Req.Host, "/"+vanityPath)
	if pkg == nil {
		proxypass.Do(c.Context)
		return
	}

	c.Reply().HTMLl("goget.html", aah.Data{
		"Host": "aahframe.work", // TODO Remove
		"Pkg":  pkg,
	})
}
