package test

import (
	ws "github.com/438577872/go-socket.io"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()
	server := ws.NewSocketServer()
	server.Install(engine)
	server.On("/", "say", func(conn *ws.SocketIOConnection, data []byte) {
		server.EmitString("/", "say", "", "hello hhh")
	})
	server.On("/", "file", func(conn *ws.SocketIOConnection, data []byte) {
		server.BroadCastBinary("file", data)
	})
	engine.Run(":5100")
}
