package websockets

import (
	"strings"

	"thumbai/app/models"

	"aahframe.work/aah"
	"aahframe.work/aah/ws"
)

// SampleWebSocket is aah's sample WebSocket implementation.
type SampleWebSocket struct {
	*ws.Context
}

// Handle method handles Text and JSON data based on Path param value.
func (w *SampleWebSocket) Handle(mode string) {
	switch strings.ToLower(mode) {
	case "text":
		w.handleTextMode()
	case "json":
		w.handleJSONMode()
	}
}

// handleTextMode method is used to communicate in Text data
func (w *SampleWebSocket) handleTextMode() {
	w.Log().Info("Handling mode: text")

	for {
		str, err := w.ReadText()
		if err != nil {
			if ws.IsDisconnected(err) {
				// WebSocket client is gone, exit here
				return
			}

			w.Log().Error(err)
			continue // we are moving on to next WS frame
		}

		// if no error then echo back Text data to WebSocket client
		if err = w.ReplyText(str); err != nil {
			w.Log().Error(err)
		}
	}
}

// handleJSONMode method is used to communicate in JSON data
func (w *SampleWebSocket) handleJSONMode() {
	w.Log().Infof("Handling mode: json")

	for {
		var greet models.Greet
		if err := w.ReadJSON(&greet); err != nil {
			if ws.IsDisconnected(err) {
				// WebSocket client is gone, exit here
				return
			}

			w.Log().Error(err)

			// JSON read error happened
			_ = w.ReplyJSON(aah.Data{"message": "invalid JSON"})

			continue // we are moving on to next WS frame
		}

		// if no error then echo back JSON data to WebSocket client
		if err := w.ReplyJSON(greet); err != nil {
			w.Log().Error(err)
		}
	}
}
