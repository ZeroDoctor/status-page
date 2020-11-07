
let socket
const connID = rand_string(12)

window.onload = function() {
	console.log("init-websocket...")
	connect()
}

function connect() {
	console.log("Attempting Connection")
	socket = new WebSocket("ws://" + document.location.host + "/ws")

	socket.addEventListener("open", function(event) {
		socket.send(JSON.stringify({
			type: "init",
			data: connID,
		}))

		console.log("Connected")
	})

	socket.addEventListener("close", function(event) {
		console.log("Socket Closed", event)
	})

	socket.addEventListener("error", function(err) {
		console.error("Websocket Error:", err)
	})

	socket.addEventListener("message", handle_message)
}

function handle_message(event) {
	const msg = JSON.parse(event.data)

	switch(msg.type) {
		
		default:
			console.log(msg)
	}
}

function rand_string(){
	return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
}
