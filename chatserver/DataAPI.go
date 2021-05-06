package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"regexp"
)

type ChatData struct {
	UserId int `json:"user_id"`
	ChannelId int `json:"channel_id"`
}

func GetUploadDataIfCanSendMessageInChannel(db *sql.DB,URL string,username string, password string) ChatData{
	results,err := db.Query("SELECT usernames.id,channels.id FROM usernames INNER JOIN channels ON usernames.channel_id=channels.id WHERE channels.url_key=? AND usernames.name=? AND usernames.password=?",URL,username,HashPassword(password))
	if err != nil {
		fmt.Println(err.Error())
	}
	var info ChatData
	for results.Next(){
		err = results.Scan(&info.UserId,&info.ChannelId)
		return info
	}
	info.UserId = -1
	info.ChannelId = -1
	return info
}

func InfoIsValid(info ChatData) bool{
	return info.UserId != -1 && info.ChannelId != -1
}

func UploadChatIfCanSend(db *sql.DB,message Message,info ChatData){
	if InfoIsValid(info){
		_,err := db.Exec("INSERT INTO chats (channel_id,message,sender_id) VALUES (?,?,?)",info.ChannelId,message.Message,info.UserId)
		if err != nil {
			panic(err.Error())
		}
	}
}



func getChannelByURL(db *sql.DB,URL string)int{
	var id *int
	err := db.QueryRow("SELECT id FROM channels WHERE url_key=?",URL).Scan(&id)
	if err != nil {
		return -1
	}
	return *id
}

func GetMessagesFromServer(db *sql.DB,URL string,count int,cipheredText string,offset int)[]Message{
	validInfo := GetServerIDIfCanView(db,URL,cipheredText)
	validMessages := make([]Message,count)
	if validInfo != -1{
		results,err := db.Query("SELECT chats.message,usernames.name FROM chats INNER JOIN usernames ON chats.sender_id=usernames.id WHERE chats.channel_id=? ORDER BY sent DESC LIMIT ? OFFSET ?",validInfo,count,offset)
		if err != nil {
			panic(err.Error())
		}
		index := 0
		var message Message
		message.Server = URL
		for results.Next(){
			err := results.Scan(&message.Message,&message.Username)
			if err != nil{
				fmt.Println(err.Error())
				break
			}
			validMessages[index] = message
			index++
		}
		fmIndex := 0
		filledMessages := make([]Message,index)
		for i := index-1;i>=0;i--{
			filledMessages[fmIndex] = validMessages[i]
			fmIndex++
		}
		validMessages = filledMessages
	}else {
		fmt.Println("Invalid credentials")
	}
	return validMessages
}

func UserInChannelWithDiffPass(db *sql.DB,channelID int,username string,password string)bool{
	var id *int
	err := db.QueryRow("SELECT id FROM usernames WHERE channel_id = ? AND name = ? AND password != ?",channelID,username,HashPassword(password)).Scan(&id)
	if err != nil{
		return false
	}
	return true
}

func CreateUserIfDoesNotExistInServer(db *sql.DB,channelId int,username string,password string) int{
	if channelId == -1 || UserInChannelWithDiffPass(db,channelId,username,password){
		return -1
	}
	return CreateUserAndGetID(db,username,password,channelId)
}

func CreateServer(db *sql.DB,URL string,name string,cleanWord string,cipheredWord string,username string,password string)bool{
	id := getChannelByURL(db,URL)
	if id != -1{
		return false
	}
	insert,err := db.Exec("INSERT INTO channels (name,url_key,clean_word,ciphered_word) VALUES (?,?,?,?)",name,URL,cleanWord,cipheredWord)
	if err != nil{
		fmt.Println("Create server error: "+err.Error())
		return false
	}else if len(username)>0 && len(password)>0{
		channelID,err := insert.LastInsertId()
		if err != nil{
			fmt.Println("Get new server ID error: "+err.Error())
			return false
		}
		viceroyID := CreateUserAndGetID(db,username,password, int(channelID))
		_,err = db.Exec("UPDATE channels SET viceroy_id = ? WHERE id = ?",viceroyID,channelID)
		if err != nil{
			fmt.Println("Error assigning viceroy: "+err.Error())
		}
	}


	return true
}

func CreateUserAndGetID(db *sql.DB,username string,password string,channelId int)int{
	insert,err := db.Exec("INSERT INTO usernames (name,password,channel_id) VALUES (?,?,?)",username,HashPassword(password),channelId)
	if err != nil{
		fmt.Println("Create user error: "+err.Error())
		return -1
	}else{
		id,err := insert.LastInsertId()
		if err != nil{
			fmt.Println("Get user ID error: "+err.Error())
		}
		return int(id)
	}

}

func GetServerIDIfCanView(db *sql.DB,URL string,cipheredWord string) int{
	var id *int
	err := db.QueryRow("SELECT id FROM channels WHERE url_key = ? AND ciphered_word = ?",URL,cipheredWord).Scan(&id)
	if err != nil{
		return -1
	}
	return *id
}

func HashPassword(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func IsViceroy(db *sql.DB,URL string,username string,password string)bool{
	var id *int
	err := db.QueryRow("SELECT usernames.id FROM usernames INNER JOIN channels ON channels.viceroy_id = usernames.id usernames.name = ? AND usernames.password = ? AND channels.name = ?",username,HashPassword(password),URL).Scan(&id)
	if err != nil{
		return false
	}
	return true
}
func LoadAllMessagesIfViceroy(db *sql.DB,URL string,username string,password string)[]Message{
	viceroy := IsViceroy(db,URL,username,password)
	validMessages := make([]Message,0)
	if viceroy{
		channelID := getChannelByURL(db,URL)
		results,err := db.Query("SELECT chats.message,usernames.name FROM chats INNER JOIN usernames ON chats.sender_id=usernames.id WHERE usernames.channel_id=? ORDER BY sent DESC",channelID)
		if err != nil {
			panic(err.Error())
		}
		var message Message
		message.Server = URL
		for results.Next(){
			err := results.Scan(&message.Message,&message.Username)
			if err != nil{
				break
			}
			validMessages = append(validMessages,message)
		}
	}
	return validMessages
}

func GetVerifyWord(db *sql.DB,URL string)string{
	var word *string
	err := db.QueryRow("SELECT clean_word FROM channels WHERE url_key = ?",URL).Scan(&word)
	if err != nil{
		fmt.Println("Couldn't get verification word from DB: "+err.Error())
		return ""
	}
	return *word
}

func ValidateCipher(db *sql.DB,URL string,cipherAttempt string)bool{
	var id *int
	err := db.QueryRow("SELECT id FROM channels WHERE url_key = ? and ciphered_word = ?",URL,cipherAttempt).Scan(&id)
	if err != nil{
		return false
	}
	return true
}

func stripString(str string)string{
	reg, err := regexp.Compile("[^a-zA-Z0-9]+")
	if err != nil {
		log.Fatal(err)
	}
	return reg.ReplaceAllString(str, "")
}

func sendMessagesToWebsocket(db *sql.DB,server string,count int,cipherText string,ws *websocket.Conn){
	err := ws.WriteJSON(GetMessagesFromServer(db, server, count, cipherText,0))
	if err != nil{
		fmt.Println("Failed to send messages from server: "+err.Error())
	}
}

func pushMessage(db *sql.DB,server string,msg MessageWithUser) bool{
	info := GetUploadDataIfCanSendMessageInChannel(db,server,msg.Username,msg.Password)
	if InfoIsValid(info){
		broadcasts[server] <- msg.Message
		go UploadChatIfCanSend(db,msg.Message,info)
	}else{
		id := getChannelByURL(db,server)
		userID := CreateUserIfDoesNotExistInServer(db,id,msg.Username,msg.Password)
		if userID != -1 {
			broadcasts[server] <- msg.Message
			info.UserId = userID
			info.ChannelId = id
			go UploadChatIfCanSend(db,msg.Message,info)
		}

	}
	return true
}

//func deleteUserName(db *sql.DB,)

//Issue here is that we only have usernames, as user doesn't have IDs.  Get userID for each username

//func DropAllAndReencryptIfViceroy(db *sql.DB,URL string,username string,password string,newMessages []DateAndMessage){
//	viceroy := IsViceroy(db,URL,username,password)
//	if viceroy{
//		channelID := getChannelByURL(db,URL)
//		_,err := db.Exec("DELETE FROM chats WHERE channel_id = ?",channelID)
//		if err != nil{
//			fmt.Println("Delete all chats error")
//		}else{
//			sqlStr := "INSERT INTO chats "
//		}
//	}
//}

//
//func getUserByName(db *sql.DB,user string)int{
//	var id *int
//	err := db.QueryRow("SELECT id FROM users WHERE name=?",user).Scan(&id)
//	if err != nil {
//		return -1
//	}
//	return *id
//}
//
//func userIsMember(db *sql.DB,user int,channel int)bool{
//	err := db.QueryRow("SELECT id FROM memberships WHERE user_id=? AND channel_id=?",user,channel)
//	if err != nil {
//		return false
//	}
//	return true
//}

//func joinChannel(db *sql.DB,user int,channel int){
//	if userIsMember(db,user,channel){
//		return
//	}
//	insert,err := db.Query("INSERT INTO memberships (user_id,channel_id) VALUES (?,?)",user,channel)
//	if err != nil {
//		panic(err.Error())
//	}
//	defer insert.Close()
//}



//func channelExists(db *sql.DB,name string) bool{
//	err := db.QueryRow("SELECT id FROM channels WHERE name=?",name)
//	if err != nil {
//		return false
//	}
//	return true
//}
//
//func userExists(db *sql.DB,name string,email string) bool{
//	err := db.QueryRow("SELECT id FROM users WHERE name=? OR email=?",name,email)
//	if err != nil{
//		return false
//	}
//	return true
//}
//
//func createChannel(db *sql.DB,name string,user int){
//	if channelExists(db,name){
//		return
//	}
//	insert,err := db.Query("INSERT INTO channels (name) VALUES (?)",name)
//	if err != nil {
//		panic(err.Error())
//	}
//	channelId := getChannelByName(db,name)
//	joinChannel(db,user, channelId)
//	defer insert.Close()
//}
//
//func createUser(db *sql.DB,name string,email string,password string) int{
//	if userExists(db,name,email){
//		return -1
//	}
//	password,_ = HashPassword(password)
//	insert,err := db.Query("INSERT INTO users (username,email,password) VALUES (?,?,?)",name,email,password)
//	if err != nil {
//		panic(err.Error())
//	}
//	defer insert.Close()
//	return getUserByName(db,name)
//}
