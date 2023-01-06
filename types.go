package ws

import "github.com/gorilla/websocket"

type Room struct {
	RoomId string `json:"roomId,omitempty"`
}
type ConnectPool map[*websocket.Conn]bool

type SocketIOConnection struct {
	Websocket *websocket.Conn
	Namespace string
	mapping   map[any]bool
}

type HelloResponse struct {
	Sid          string   `json:"sid"`
	Upgrades     []string `json:"upgrades"`
	PingTimeout  int      `json:"pingTimeout"`
	PingInterval int      `json:"pingInterval"`
}
type RoomType = map[string]SocketSet
type SocketSet = map[*SocketIOConnection]bool
type SocketServer struct {
	eventHandler map[string]map[string]HandleFunction
	namespaceMap map[string]RoomType
}

type HandleFunction = func(conn *SocketIOConnection, data []byte)
type Message struct {
	Code      int
	EventName string
	Namespace string
	Data      []byte
}
