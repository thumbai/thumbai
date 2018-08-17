package controllers

import (
	"aahframework.org/aah.v0"
)

// AppController struct application controller
type AppController struct {
	*aah.Context
}

// Index method is application's home page.
func (c *AppController) Index() {
	c.Reply().Text("Coming soon...")
	// data := aah.Data{
	//   "Greet": models.Greet{
	//     Message: "Welcome to aah framework - Web Application",
	//   },
	// }

	// c.Reply().Ok().HTML(data)
}
