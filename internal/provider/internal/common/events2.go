package common

import (
	"fmt"
	"log"

	"golang.org/x/net/websocket"
)

// This example demonstrates a trivial client.
func WSSConnect() {
	origin := "http://ipm-eval4.westus3.cloudapp.azure.com/"
	url := "ws://ipm-eval4.westus3.cloudapp.azure.com/api/v1/ws/events/fb4507b8-b6db-4838-8012-60edbb2b10da"
	ws, err := websocket.Dial(url, "", origin)
	if err != nil {
		log.Println("Cannot connect to websocket: ws://ipm-eval4.westus3.cloudapp.azure.com/api/v1/ws/events/fb4507b8-b6db-4838-8012-60edbb2b10da")
		log.Fatal(err)
	}
	var msg = make([]byte, 512)
	var n int
	if n, err = ws.Read(msg); err != nil {
		log.Println("Cannot read to websocket: ws://ipm-eval4.westus3.cloudapp.azure.com/api/v1/ws/events/fb4507b8-b6db-4838-8012-60edbb2b10da")
		log.Fatal(err)
	}
	fmt.Printf("Received: %s.\n", msg[:n])
}
