const zmq = require('zeromq/v5-compat')
const responder = zmq.socket('rep')


responder.bind('tcp://\*:60123', err => {
    if (err) throw err
    console.log("готов играть")
})

let min
let max
let ans

function getRndNum(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min;
}

responder.on('message', data => {
    const msg = JSON.parse(data)
    if (msg.hint == "more") {
        max = ans
    } else if (msg.hint == "less") {
        min = ans
    } else if (msg.hint == "число угадано") {
        console.log(msg.hint)
        return
    } else {
        min = msg.min
        max = msg.max
    }
    ans = getRndNum(min,max)
    responder.send(ans)
    console.log(msg)
})