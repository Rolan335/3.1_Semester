const zmq = require('zeromq/v5-compat')
const requester = zmq.socket('req')

const min = parseInt(process.argv[2])
const max = parseInt(process.argv[3])

function getRndNum(min, max) {
    return Math.floor(Math.random() * (max - min + 1)) + min;
}
const number = getRndNum(min, max)
console.log(number)

requester.send(`{"min":${min},"max":${max}}`)

requester.on('message', data =>{
    const ans = JSON.parse(data)
    if (ans > number){
        requester.send("{\"hint\": \"more\"}")
    } else if (ans < number){
        requester.send("{\"hint\": \"less\"}")
    } else{
        requester.send("{\"hint\": \"число угадано\"}")
    }
    console.log(ans)
})

requester.connect("tcp://localhost:60123")