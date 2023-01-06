package socketio

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"net/http"
	"strconv"
	"time"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func NewSocketServer() *SocketServer {
	var s = new(SocketServer)
	s.eventHandler = make(map[string]map[string]HandleFunction)
	s.namespaceMap = make(map[string]RoomType)
	return s
}

func (this *SocketServer) On(namespace, eventName string, fn HandleFunction) {
	if _, ok := this.eventHandler[namespace]; !ok {
		this.eventHandler[namespace] = make(map[string]HandleFunction)
	}
	this.eventHandler[namespace][eventName] = fn
}

func (this *SocketServer) JoinRoom(ws *Connection, namespace, roomId string) {
	_, ok := this.namespaceMap[namespace]
	if !ok {
		this.namespaceMap[namespace] = make(RoomType)
	}
	_, ok = this.namespaceMap[namespace][roomId]
	if !ok {
		this.namespaceMap[namespace][roomId] = make(SocketSet)
	}
	this.namespaceMap[namespace][roomId][ws] = true
	ws.mapping[&this.namespaceMap] = true
}

// 紫砂了
func (this *SocketServer) destroy(ws *Connection) {
	for t := range ws.mapping {
		m := t.(*map[*Connection]bool)
		delete(*m, ws)
	}
}

func (this *SocketServer) emit(namespace, roomId string, code int, data []byte) {
	room, ok := this.namespaceMap[namespace]
	if !ok {
		return
	}
	socketSet, ok := room[roomId]
	if !ok {
		return
	}

	for conn := range socketSet {
		_ = conn.Websocket.WriteMessage(code, data)
	}
}

//func (this *SocketServer) EmitJSON(nsp, event, roomId string, data any) {
//
//}

func (this *SocketServer) EmitBinary(namespace, event, roomId string, data []byte) {
	var payload = make([]any, 0, 2)
	payload = append(payload, event, json.RawMessage(`{"_placeholder":true,"num":0}`))
	b, _ := json.Marshal(payload)
	sendMessage := append([]byte(strconv.Itoa(BinaryType)), '-')
	sendMessage = append(sendMessage, b...)

	this.emit(namespace, roomId, 1, sendMessage)
	this.emit(namespace, roomId, 2, data)

}

func (this *SocketServer) BroadCastBinary(event string, data []byte) {
	this.EmitBinary("/", event, "", data)
}

func (this *SocketServer) BroadCastString(event string, data string) {
	this.EmitString("/", event, "", data)
}

func (this *SocketServer) EmitString(namespace, event, roomId string, data string) {
	var payload = make([]string, 0, 2)
	payload = append(payload, event, data)
	b, _ := json.Marshal(payload)
	sendMessage := append([]byte(strconv.Itoa(TextType)))
	if namespace != "/" && namespace != "" {
		sendMessage = append(sendMessage, []byte(namespace)...)
		sendMessage = append(sendMessage, ',')
	}
	sendMessage = append(sendMessage, b...)

	this.emit(namespace, roomId, 1, sendMessage)
}
func (this *SocketServer) generateName() string {
	return uuid.New().String()
}

func (this *SocketServer) helloCallback(connection *Connection) string {
	name := this.generateName()
	var hello = HelloResponse{
		Sid:          name,
		Upgrades:     make([]string, 0),
		PingTimeout:  20000,
		PingInterval: 25000,
	}
	marshal, _ := json.Marshal(hello)
	_ = connection.Websocket.WriteMessage(1, append([]byte{'0' + HelloType}, marshal...))
	return name
}

func (this *SocketServer) parseRoomNamespace(connection *Connection) {
	_, data, _ := connection.Websocket.ReadMessage()
	connection.Namespace = "/"
	if n := len(data); n > 2 {
		connection.Namespace = string(data[2 : n-1])
	}
	_ = connection.Websocket.WriteMessage(1, append(data, []byte(fmt.Sprintf(`{"sid":"%s"}`, this.generateName()))...))
}

func escape(data []byte) []byte {
	n := len(data)
	return data[1 : n-1]
}

func (this *SocketServer) parseMessage(connection *Connection) (msg Message, err error) {
	_, data, err := connection.Websocket.ReadMessage()
	if err != nil {
		return
	}
	var payload []byte
	var n = len(data)
	var end int
	if len(data) == 1 && data[0]-'0' == PingIn {
		msg.Code = PingIn
		return
	}
	for i := 0; i < n; i++ {
		var char = data[i]
		if char <= '9' && char >= '0' {
			msg.Code *= 10
			msg.Code += int(char - '0')
		} else {
			end = i
			break
		}
	}
	for i := end; i < n; i++ {
		var char = data[i]
		if char == '[' {
			payload = data[i:]
			break
		} else {
			msg.Namespace += string(char)
		}
	}
	if n := len(msg.Namespace); n > 0 && msg.Namespace[0] == '-' {
		msg.Namespace = msg.Namespace[1:]
	}
	if n := len(msg.Namespace); n > 0 && msg.Namespace[n-1] == ',' {
		msg.Namespace = msg.Namespace[:n-1]
	}

	if msg.Namespace == "" {
		msg.Namespace = "/"
	}

	var msgs = make([]json.RawMessage, 2)

	_ = json.Unmarshal(payload, &msgs)

	msg.EventName = string(escape(msgs[0]))
	if msg.Code == BinaryType {
		_, msg.Data, _ = connection.Websocket.ReadMessage()
	}
	if msg.Code == TextType {
		msg.Data = escape(msgs[1])
	}
	return

}

func (this *SocketServer) getConnection(c *gin.Context) (*Connection, error) {
	upgrade, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	return &Connection{
		Websocket: upgrade,
		mapping:   map[any]bool{},
	}, err
}

func (this *SocketServer) ws(c *gin.Context) {

	conn, err := this.getConnection(c)
	if err != nil {
		return
	}

	this.helloCallback(conn)
	this.parseRoomNamespace(conn)
	this.JoinRoom(conn, conn.Namespace, "")

	for {
		message, err := this.parseMessage(conn)

		if message.Code == PingIn {
			continue
		}
		if err != nil {
			break
		}

		if _, ok := this.eventHandler[message.Namespace]; ok {
			if _, ok = this.eventHandler[message.Namespace][message.EventName]; ok {
				this.eventHandler[message.Namespace][message.EventName](conn, message.Data)
			}
		}
	}
}

func (this *SocketServer) Install(app *gin.Engine) {
	app.GET("/socket.io/", this.ws)
	go func() {
		for {
			var data = []byte{'0' + PingBack}
			time.Sleep(20 * time.Second)
			for _, s := range this.namespaceMap {
				for conn, _ := range s[""] {
					conn.Websocket.WriteMessage(1, data)
				}
			}
		}
	}()
}
