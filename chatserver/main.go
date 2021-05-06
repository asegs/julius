package main

//
//import (
//	"github.com/gorilla/websocket"
//	"log"
//	"net/http"
//)
//
//var clients = make(map[*websocket.Conn]bool)
//var broadcast = make(chan Message)
//
//var upgrader = websocket.Upgrader{}
//
//var id = 0
//
//type Message struct {
//	Username string `json:"username"`
//	Message string `json:"message"`
//}
//
//type IDMessage struct {
//	Username string `json:"username"`
//	Message string `json:"message"`
//	ID int `json:"id"`
//}
//
//func handleConnections(w http.ResponseWriter,r *http.Request){
//	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
//	ws,err := upgrader.Upgrade(w,r,nil)
//	if err != nil{
//		log.Fatal(err)
//	}
//
//	defer ws.Close()
//	clients[ws] = true
//	log.Println("user connected")
//	for {
//		var msg Message
//		err:= ws.ReadJSON(&msg)
//		if err != nil{
//			log.Printf("error: %v",err)
//			delete(clients,ws)
//			break
//		}
//		broadcast <- msg
//	}
//}
//
//func handleMessages(){
//	for {
//		msg := <-broadcast
//		fmtMsg := IDMessage{
//			Username: msg.Username,
//			Message:  msg.Message,
//			ID:       id,
//		}
//		for client := range clients{
//			err := client.WriteJSON(fmtMsg)
//			if err != nil{
//				log.Printf("error: %v",err)
//				client.Close()
//				delete(clients,client)
//			}
//		}
//		id++
//	}
//}
//
//func main(){
//	fs := http.FileServer(http.Dir("../public"))
//	http.Handle("/",fs)
//	http.HandleFunc("/ws",handleConnections)
//	go handleMessages()
//	log.Println("http server started on :3005")
//	err := http.ListenAndServe(":3005",nil)
//	if err != nil{
//		log.Fatal("ListenAndServe: ",err)
//	}
//}