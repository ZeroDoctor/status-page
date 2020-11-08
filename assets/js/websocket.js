
let socket
let log_container
const connID = rand_string(12)

const log_types = {
	"info":["bg-secondary", "text-primary"],
	"warn":["bg-warning", "text-light"],
	"error":["bg-error", "text-light"],
}

window.onload = function() {
	console.log("init-websocket...")
	connect()

	log_container = document.getElementById("logs")
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
		case "log":
			add_log(msg.data)
		break;
		
		default:
			console.log(msg)
	}
}

function add_log(data) {

	type = data["type"]
	bg_color = log_types[type][0]
	text_color = log_types[type][1]

	file_name = data["file_name"]
	func_name = data["func_name"]
	line_number = data["line_number"]

	let html = `
	<div class="accordion">
		<input type="checkbox" id="accordion-`+data["index"]+data["app_id"]+`" name="accordion-checkbox" hidden>

		<div class="columns col-oneline `+bg_color+` `+text_color+`">
			<div class="column col-8">
				<label class="accordion-header" for="accordion-`+data["index"]+data["app_id"]+`">
					<i class="icon icon-arrow-right mr-1"></i>
					`+type+`
				</label>
			</div>
			<div class="column col-4">
				<p> `+file_name+` | `+func_name+` | `+line_number+` </p>
			</div>
		</div>

		<div class="accordion-body">
			<div class="tile p-2 bg-gray">
				<div class="tile-content">
					<p class="tile-title"> `+data["msg"]+` </p>
				</div>
			</div>
		</div>
	</div>
	` 

	var new_content = document.createElement("div")
	new_content.innerHTML = html.trim()

	log_container.appendChild(new_content)
}

function rand_string(){
	return Math.random().toString(36).substring(2, 15) + Math.random().toString(36).substring(2, 15);
}
