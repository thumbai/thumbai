package v1

import (
	"fmt"

	"thumbai/app/models"

	"aahframe.work/aah"
	"aahframe.work/aah/ahttp"
)

var values = make(map[string]*models.Value)

// ValueController is kickstart sample for API implementation.
type ValueController struct {
	*aah.Context
}

// List method returns all the values.
func (c *ValueController) List() {
	c.Reply().Ok().JSON(values)
}

// Index method returns value for given key.
// If key not found then returns 404 NotFound error.
func (c *ValueController) Index(key string) {
	if val, found := values[key]; found {
		c.Reply().Ok().JSON(val)
		return
	}

	c.Reply().NotFound().JSON(aah.Data{
		"message": "Value not exists",
	})
}

// Create method creates new entry in the values map with given payload.
// If key already exists then returns 409 Conflict error.
func (c *ValueController) Create(val *models.Value) {
	if _, found := values[val.Key]; found {
		c.Reply().Conflict().JSON(aah.Data{
			"message": "Key already exists",
		})
		return
	}

	// Add it to values map
	values[val.Key] = val
	newResourceURL := fmt.Sprintf("%s:%s", c.Req.Scheme, c.RouteURL("value_get", val.Key))
	c.Reply().Created().
		Header(ahttp.HeaderLocation, newResourceURL).
		JSON(aah.Data{
			"key": val.Key,
		})
}

// Update method updates value entry on map for given key and Payload.
// If key not exists then returns 400 BadRequest error.
func (c *ValueController) Update(key string, val *models.Value) {
	if r, found := values[key]; found {
		r.Value = val.Value
		values[key] = r
		c.Reply().Ok().JSON(aah.Data{
			"message": "Value updated successfully",
		})
		return
	}

	c.Reply().BadRequest().JSON(aah.Data{
		"message": "Invalid input",
	})
}

// Delete method deletes value for given key.
// If key not exists then returns 400 BadRequest error.
func (c *ValueController) Delete(key string) {
	if _, found := values[key]; found {
		delete(values, key)
		c.Reply().Ok().JSON(aah.Data{
			"message": "Value deleted successfully",
		})
		return
	}

	c.Reply().BadRequest().JSON(aah.Data{
		"message": "Invalid input",
	})
}
