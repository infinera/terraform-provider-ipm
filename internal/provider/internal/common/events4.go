package common

import (
	"log"
	//"net/url"
	//"os"
	//"os/signal"

	//"github.com/gorilla/websocket"
	"golang.org/x/net/websocket"
)

func WSSConnect3() {
	//interrupt := make(chan os.Signal, 1)
	//signal.Notify(interrupt, os.Interrupt)
	uri := "ws://ipm-eval4.westus3.cloudapp.azure.com/api/v1/ws/events/a57bc251-447c-4aae-8a48-5e52854f4230"


// create connection
// schema can be ws:// or wss://
// host, port – WebSocket server
c, err := websocket.Dial(uri, "", "")
	if err != nil {
		log.Fatal("#### Dial WSS failed:"+ uri , err)
	}
	defer c.Close()

	log.Println("*** Dial WSS Success:"+ uri)

}
