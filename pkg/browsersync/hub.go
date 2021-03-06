package browsersync

type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Broadcast a message to all registered clients
	broadcast chan []byte

	// Register a new client
	register chan *Client

	// Unregister a client
	unregister chan *Client
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) listen() {
	for {
		select {
		case client := <-h.register:
			h.registerClient(client)
		case client := <-h.unregister:
			h.unregisterClient(client)
		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

func (h *Hub) registerClient(client *Client) {
	h.clients[client] = true
}

func (h *Hub) unregisterClient(client *Client) {
	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.inboundMessage)
	}
}

func (h *Hub) broadcastMessage(message []byte) {
	for client := range h.clients {
		select {
		case client.inboundMessage <- message:
		default:
			close(client.inboundMessage)
			delete(h.clients, client)
		}
	}
}
