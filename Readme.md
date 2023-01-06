实现方式与官方不同

目前只支持ws方式
所以需要手动指定
```ts

const io = socket('ws://localhost:' + port + namespace, { transports: [ 'websocket' ] })
```

前端连接时如果不指定namespace, 那么默认的namespace='/'

所有连接都会存在 roomId="" 的房间里, 如果需要额外的房间则需要调用joinRoom

namespaceMap 大致可以理解成 map<namespace, map<roomId, set<connection> > >

BroadCast 发送给所有namespace='/' roomId=''的用户

Emit发送给指定namespace和roomId的用户
```go
EmitString(namespace, eventName, roomId, stringData)
EmitBinary(namespace, eventName, roomId, byteData)
```
### 使用方法
> 创建一个gin工程, 调用NewSocketServer, 把gin注册到server即可

> server.On注册相关事件
```go
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

```