package browsersync

// Hub is a client switch that broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Broadcast a message to all registered clients
	broadcast chan []byte

	// Register a new client
	register chan *Client

	stopListening chan bool
}

// NewHub creates a new Hub
func NewHub() *Hub {
	return &Hub{
		clients:       make(map[*Client]bool),
		broadcast:     make(chan []byte),
		register:      make(chan *Client),
		stopListening: make(chan bool, 1),
	}
}

func (h *Hub) stop() {
	for client := range h.clients {
		h.unregisterClient(client)
	}
	h.stopListening <- true
}

func (h *Hub) listen() {
OuterLoop:
	for {
		select {
		case <-h.stopListening:
			break OuterLoop
		case client := <-h.register:
			h.registerClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
	close(h.stopListening)

}

func (h *Hub) registerClient(client *Client) {
	h.clients[client] = true
}

func (h *Hub) unregisterClient(client *Client) {
	delete(h.clients, client)
	client.close()
}

func (h *Hub) broadcastMessage(message []byte) {
	for client := range h.clients {
		select {
		case client.outboundMessage <- message:
		default:
			h.unregisterClient(client)
		}
	}
}
