package eventsquery

import (
	"context"

	"github.com/gorilla/websocket"

	"pocket/pkg/client"
)

var _ client.Dialer = &websocketDialer{}

// websocketDialer implements the Dialer interface using the gorilla websocket
// transport implementation.
type websocketDialer struct{}

// NewWebsocketDialer creates a new websocketDialer.
func NewWebsocketDialer() client.Dialer {
	return &websocketDialer{}
}

// DialContext implements the respective interface method using the default gorilla
// websocket dialer.
func (wsDialer *websocketDialer) DialContext(
	ctx context.Context,
	urlString string,
) (client.Connection, error) {
	// TODO_IMPROVE: check http response status and potential err
	// TODO_TECHDEBT: add test coverage and ensure support for a 3xx responses
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, urlString, nil)
	if err != nil {
		return nil, err
	}
	return &websocketConn{conn: conn}, nil
}
