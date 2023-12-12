package common

import (
	"log"
	"net/url"
	//"os"
	//"os/signal"

	"github.com/gorilla/websocket"
)

func WSSConnect2() {
	//interrupt := make(chan os.Signal, 1)
	//signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: "ipm-eval4.westus3.cloudapp.azure.com", Path: "/api/v1/ws/events/a57bc251-447c-4aae-8a48-5e52854f4230"}
	log.Printf("connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial("ws://ipm-eval4.westus3.cloudapp.azure.com/api/v1/ws/events/a57bc251-447c-4aae-8a48-5e52854f4230", nil)
	if err != nil {
		log.Fatal("#### Dial WSS failed:", err)
	}
	defer c.Close()

	log.Println("*** Dial WSS Success:"+ "//ipm-eval4.westus3.cloudapp.azure.com/api/v1/ws/events/fb4507b8-b6db-4838-8012-60edbb2b10da")

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

}
