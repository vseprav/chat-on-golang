package main

import (
	"net/http"
	"time"
	"html/template"
	"golang.org/x/net/websocket"
)

const templatesPath string = "templates/"

var (
	msgChan             = make(chan string, 100)
	clientDisconnects   = make(chan time.Time, 100)
	clientRequests      = make(chan *ClientRequest, 100)
)

type (

	ClientRequest struct {
		clientKey time.Time
		conn   *websocket.Conn
	}
)

func main() {
	http.HandleFunc("/", handleIndexPage)
	go router()
	http.Handle("/chat", websocket.Handler(chatServer))
	err := http.ListenAndServe(":8081", nil)
	if err != nil {
		panic("ListenAndServe: " + err.Error())
	}
}

func chatServer(ws *websocket.Conn) {
	var clientKey = time.Now()
	clientRequests <- &ClientRequest{clientKey, ws}

	var message string
	for {
		err := websocket.Message.Receive(ws, &message)
		if err == nil {
			msgChan <- message
		}
	}
	defer func() {
		clientDisconnects <- clientKey
	}()
}

func router() {
	var connections = make(map[time.Time]*websocket.Conn)

	for {
		select {
		case req := <-clientRequests:
			connections[req.clientKey] = req.conn
		case msg := <-msgChan:
			for _, con := range connections {
				websocket.Message.Send(con, msg)
			}
		case clientKey := <-clientDisconnects:
			delete(connections, clientKey)
		}
	}
}

func handleIndexPage(w http.ResponseWriter, r *http.Request)  {
	indexPage := template.Must(template.ParseFiles(templatesPath + "index.html"))

	if err := indexPage.ExecuteTemplate(w, "index.html", nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
