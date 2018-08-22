package models

// Value model is used by ValueController
type Value struct {
	Key   string      `json:"key" validate:"required"`
	Value interface{} `json:"value" validate:"required"`
}

// Greet ...
type Greet struct {
	Message string `json:"message"`
}
