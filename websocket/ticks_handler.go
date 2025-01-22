package websocket

import (
	"auto_trade/ticks"
	"log"
	"net/http"
)

type TicksHub struct {
	*Hub
	generator *ticks.TickGenerator
}

func NewTicksHub() *TicksHub {
	return &TicksHub{
		Hub:       NewHub(),
		generator: ticks.NewTickGenerator(),
	}
}

func (h *TicksHub) Run() {
	// Start the base hub
	go h.Hub.Run()

	// Start tick generation
	go h.generator.StartGeneration(h.broadcast)
}

// ServeTicksWs handles websocket requests for market data ticks
func ServeTicksWs(hub *TicksHub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		hub:  hub.Hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	client.hub.register <- client

	go client.writePump()
}
