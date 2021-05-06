package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)
var clients = make(map[string]map[*websocket.Conn]bool)
var broadcasts = make(map[string]chan Message)
var db = openConnection()

type Message struct {
	Username string `json:"username"`
	Message string `json:"message"`
	Server string `json:"server"`
}

type DateAndMessage struct {
	Message Message `json:"message"`
	Date string `json:"date"`
}

type NewServer struct {
	URL string `json:"url"`
	Name string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	CleanText string `json:"clean_text"`
	CipherText string `json:"cipher_text"`
}

type ValidatorCombo struct {
	URL string `json:"url"`
	CipherText string `json:"cipher_text"`
}

type UserInfo struct {
	URL string `json:"url"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type MessageWithUser struct {
	Message Message `json:"message"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type URLKey struct {
	URL string `json:"url"`
}

type MessageRequest struct {
	Server string `json:"server"`
	CipherText string `json:"cipher_text"`
	Offset int `json:"offset"`
}

var upgrader = websocket.Upgrader{}

func trimmer(s string,l int) string{
	if len(s)>l{
		return s[0:l]
	}
	return s
}

func handleConnections(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin","*")
	w.Header().Set("Access-Control-ALlow-Headers","*")
	w.Header().Set("Access-Control-Allow-Methods","*")
	if r.Method == http.MethodOptions{
		w.WriteHeader(http.StatusOK)
		return
	}
	upgrader.CheckOrigin = func(r *http.Request) bool {return true}
	ws,err := upgrader.Upgrade(w,r,nil)
	w.Header().Set("Content-Type","application/json")
	var Combo ValidatorCombo
	vars := mux.Vars(r)
	URL := vars["server"]
	validation := vars["code"]
	Combo.URL = URL
	Combo.CipherText = validation
	validCipher := ValidateCipher(db,Combo.URL,Combo.CipherText)
	if validCipher {
		server := Combo.URL
		if _, ok := clients[server]; !ok {
			fmt.Println("New chat server group spawned at " + server)
			clients[server] = make(map[*websocket.Conn]bool)
			broadcasts[server] = make(chan Message)
			go handleMessages(server)
		}
		clients[server][ws] = true
		go sendMessagesToWebsocket(db, server, 25, Combo.CipherText,ws)
		for {
			var msg MessageWithUser
			err := ws.ReadJSON(&msg)
			if err != nil {
				log.Printf("error: %v", err)
				delete(clients[server], ws)
				break
			}
			msg.Username = trimmer(msg.Username,32)
			msg.Password = trimmer(msg.Password,64)
			msg.Message.Username = trimmer(msg.Message.Username,32)
			msg.Message.Message = trimmer(msg.Message.Message,2048)
			go pushMessage(db,server,msg)

		}
		if err != nil {
			log.Println(err.Error())
		}
	}

}


func handleMessages(server string){
	for {
		msg := <-broadcasts[server]
		for client := range clients[server]{
			err := client.WriteJSON(msg)
			if err != nil{
				log.Printf("error: %v",err)
				client.Close()
				delete(clients[server],client)
			}
		}
	}
}

func createServerAtURL(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin","*")
	w.Header().Set("Access-Control-ALlow-Headers","*")
	w.Header().Set("Access-Control-Allow-Methods","*")
	if r.Method == http.MethodOptions{
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type","application/json")
	var newServer NewServer
	err := json.NewDecoder(r.Body).Decode(&newServer)
	if err != nil{
		fmt.Println("Malformed JSON to create server: "+err.Error())
	}
	newServerSuccess := CreateServer(db,newServer.URL,newServer.Name,newServer.CleanText,stripString(newServer.CipherText),newServer.Username,newServer.Password)
	err = json.NewEncoder(w).Encode(newServerSuccess)
}


func accessServerAtURL(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin","*")
	w.Header().Set("Access-Control-ALlow-Headers","*")
	w.Header().Set("Access-Control-Allow-Methods","*")
	if r.Method == http.MethodOptions{
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type","application/json")
	var URL URLKey
	err := json.NewDecoder(r.Body).Decode(&URL)
	fmt.Println("Server "+URL.URL+" accessed")
	if err != nil{
		fmt.Println("Malformed JSON to get verification word: "+err.Error())
	}
	channelID := getChannelByURL(db,URL.URL)
	if channelID == -1{
		_ = json.NewEncoder(w).Encode(false)
	}else{
		_ = json.NewEncoder(w).Encode(GetVerifyWord(db,URL.URL))
	}
}

func getMessagesAtServerOffset(w http.ResponseWriter, r *http.Request){
	w.Header().Set("Access-Control-Allow-Origin","*")
	w.Header().Set("Access-Control-ALlow-Headers","*")
	w.Header().Set("Access-Control-Allow-Methods","*")
	if r.Method == http.MethodOptions{
		w.WriteHeader(http.StatusOK)
		return
	}
	w.Header().Set("Content-Type","application/json")
	var Request MessageRequest
	err := json.NewDecoder(r.Body).Decode(&Request)
	if err != nil{
		fmt.Println("Malformed JSON to get older messages: "+err.Error())
	}
	messages := GetMessagesFromServer(db,Request.Server,25,Request.CipherText,Request.Offset)
	_ = json.NewEncoder(w).Encode(messages)
}




func main() {
	r := mux.NewRouter()
	r.HandleFunc("/ws/{server}/{code}", handleConnections)
	r.HandleFunc("/create",createServerAtURL).Methods(http.MethodPost,http.MethodOptions)
	r.HandleFunc("/access",accessServerAtURL).Methods(http.MethodPost,http.MethodOptions)
	r.HandleFunc("/older",getMessagesAtServerOffset).Methods(http.MethodPost,http.MethodOptions)
	r.Use(mux.CORSMethodMiddleware(r))
	log.Println("http server started on :3005")
	err := http.ListenAndServe(":3005",r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}