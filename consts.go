package socketio

const (
	TextType    = 42  // 纯文本类型
	ConnectType = 40  // 握手的类型 后面跟一个room
	BinaryType  = 451 // 二进制-
	HelloType   = 0   // 握手0
	PingBack    = 2
	PingIn      = 3
)
