package lib

import (
	"fmt"
	"net/http"
	"time"
)

// A Broker holds open client connections,
// listens for incoming events on its Notifier channel
// and broadcast event data to all registered connections
type Broker struct {
	// Events are pushed to this channel by the main events-gathering routine
	Notifier chan []byte

	// New client connections
	newClients chan chan []byte

	// Closed client connections
	closingClients chan chan []byte

	// Client connections registry
	clients map[chan []byte]bool
}

func NewStreamServer() (broker *Broker) {
	// Instantiate a broker
	broker = &Broker{
		Notifier:       make(chan []byte, 1),
		newClients:     make(chan chan []byte),
		closingClients: make(chan chan []byte),
		clients:        make(map[chan []byte]bool),
	}

	// Set it running - listening and broadcasting events
	go broker.listen()

	return
}

func (broker *Broker) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	// Make sure that the writer supports flushing.
	flusher, ok := rw.(http.Flusher)
	if !ok {
		http.Error(rw, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}

	rw.Header().Set("X-Accel-Buffering", "no")
	rw.Header().Set("Content-Type", "text/event-stream")
	rw.Header().Set("Cache-Control", "no-cache")
	rw.Header().Set("Connection", "keep-alive")
	rw.Header().Set("Access-Control-Allow-Origin", "*")

	// Each connection registers its own message channel with the Broker's connections registry
	messageChan := make(chan []byte)

	// Signal the broker that we have a new connection
	broker.newClients <- messageChan

	// Remove this client from the map of connected clients
	// when this handler exits.
	defer func() {
		broker.closingClients <- messageChan
	}()

	// notify on connection closed
	notify := req.Context().Done()

	// Listen to connection close and un-register messageChan
	go func() {
		<-notify
		broker.closingClients <- messageChan
	}()

	// send wait
	go func(rw http.ResponseWriter) {
		time.Sleep(time.Millisecond * 100)
		_, _ = rw.Write([]byte("retry: 10000\n\n"))
		flusher.Flush()
	}(rw)

	// send periodic ping
	go func(rw http.ResponseWriter) {
		for {
			select {
			// close
			case <-notify:
				return
			// ping
			default:
				time.Sleep(time.Millisecond * 10000)
				_, _ = rw.Write([]byte("event: ping\n"))
				_, _ = rw.Write(fmt.Appendf(nil, "data: {\"time\":%d}\n\n", time.Now().Unix()))
				flusher.Flush()
			}
		}
	}(rw)

	for {
		// Write to the ResponseWriter
		// Server Sent Events compatible
		_, _ = fmt.Fprintf(rw, "data: %s\n\n", <-messageChan)

		// Flush the data immediately instead of buffering it for later.
		flusher.Flush()
	}
}

func (broker *Broker) listen() {
	for {
		select {
		case s := <-broker.newClients:
			// A new client has connected.
			// Register their message channel
			broker.clients[s] = true
			//log.Printf("Client added. %d registered clients", len(broker.clients))
		case s := <-broker.closingClients:
			// A client has detached and we want to
			// stop sending them messages.
			delete(broker.clients, s)
			//log.Printf("Removed client. %d registered clients", len(broker.clients))
		case event := <-broker.Notifier:
			// We got a new event from the outside!
			// Send event to all connected clients
			for clientMessageChan := range broker.clients {
				clientMessageChan <- event
			}
		}
	}
}
