import socket from 'socket.io-client'

// console.log(location.pathname.substring(1))
const port = 5100

// const io = socket('ws://localhost:5000', { transports: [ 'websocket' ] })
const io = socket('ws://localhost:' + port + '', { transports: [ 'websocket' ] })

document.querySelector('button')!.onclick = () => {
    io.emit('say', 'hello')
}
io.on('say', (s)=>{
    console.log('say back ',s)
})