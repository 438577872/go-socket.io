package test

import (
	sio "github.com/438577872/go-socket.io"
	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()
	server := sio.NewSocketServer()
	server.Install(engine)
	server.On("/", "say", func(conn *sio.Connection, data []byte) {
		server.EmitString("/", "say", "", "hello hhh")
	})
	server.On("/", "file", func(conn *sio.Connection, data []byte) {
		server.BroadCastBinary("file", data)
	})
	engine.Run(":5100")
}
