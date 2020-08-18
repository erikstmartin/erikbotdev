<script>
	import Alert from "./components/Alert.svelte";
	import Chat from "./components/Chat.svelte";

    let alert;
	let messages = [];

	function appendChat(msg) {
        // TODO: minimize max size of this array
        let start = messages.length - 20;
        if(start < 0){
            start = 0;
        }
		messages = [...messages.slice(start,messages.length), msg];
    }

    let playlist = [];
    let audioPlayer;

    function soundPlayer(){
       if(playlist.length == 0){
           return
       }

        
       if(audioPlayer !== null && audioPlayer !== undefined && audioPlayer.duration !== NaN && audioPlayer.currentTime < audioPlayer.duration){
           return
       }

        audioPlayer = new Audio(playlist.shift());
        audioPlayer.play();
    }

    function processMessage(msg){
        if(msg.type == "bot.PlaySoundMessage"){
            // TODO: We should be able to play any sound type
            playlist.push("/media/" + msg.message.sound + ".wav")
        } else if(msg.type == "http.ChatMessage") {
            appendChat(msg.message);
        } else if(msg.type == 'http.RaidMessage') {
            alert = msg.message.message;
        }
    }

    // TODO: DELETE ME!!! DO IT! DO IT NOW!!
    window.processMessage = processMessage;

    setInterval(soundPlayer, 1000);

    function startWebSocket(){
        let conn = new WebSocket("ws://" + document.location.host + "/ws");
        conn.onclose = function (evt) {
            setTimeout(startWebSocket, 5000);
        };
        conn.onmessage = function (evt) {
            var messages = evt.data.split('\n');

            for (var i = 0; i < messages.length; i++) {
                if(messages[i].length > 0){
                    var msg = JSON.parse(messages[i]);
                    console.log(msg);
                    processMessage(msg);
                }
            }
        };
    }

	if (window["WebSocket"]) {
        startWebSocket();
    } else {
        document.body.innerHTML = "<b>Your browser does not support WebSockets.</b>";
    }
</script>
<main>
    <Alert message={alert} />
	<Chat {messages} />
</main>

<style>
	main {
		background: url('/assets/overlay.png') fixed center no-repeat;
		background-size: cover;
		width: 100%;
		height: 100%;
	}
</style>
