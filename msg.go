package hitbot

import (
	"encoding/json"
	"log"
	"strings"

	"github.com/gorilla/websocket"
)

type outMessage struct {
	Name string `json:"name"`
	Args []arg  `json:"args"`
}

type inMessage struct {
	Name string   `json:"name"`
	Args []string `json:"args"`
}

type arg struct {
	Method string      `json:"method"`
	Params interface{} `json:"params"`
}

type chatArg struct {
	Method string     `json:"method"`
	Params ChatParams `json:"params"`
}

//ChatParams is a struct of params provided for chat messages.
type ChatParams struct {
	Channel      string `json:"channel"`
	Name         string `json:"name"`
	NameColor    string `json:"nameColor"`
	Text         string `json:"text"`
	Time         int    `json:"time"`
	Role         string `json:"role"`
	IsFollower   bool   `json:"isFollower"`
	IsSubscriber bool   `json:"isSubscriber"`
	IsOwner      bool   `json:"isOwner"`
	IsStaff      bool   `json:"isStaff"`
	IsCommunity  bool   `json:"isCommunity"`
	Media        bool   `json:"media"`
	Image        string `json:"image"`
	Buffer       bool   `json:"buffer"`
	BufferSent   bool   `json:"buffersent"`
}

type joinChannelParams struct {
	Channel string `json:"channel"`
	Name    string `json:"name"`
	Token   string `json:"token"`
	IsAdmin bool   `json:"isAdmin"`
}

type msgParams struct {
	Channel   string `json:"channel"`
	Name      string `json:"name"`
	NameColor string `json:"nameColor"`
	Text      string `json:"text"`
}

//JoinChannel joins channel specified in the parameter.
func (bot *Hitbot) JoinChannel(channel string) {
	bot.joinChannel(channel)
	if bot.verbose {
		log.Printf("Attempted login to \"%v\"...", channel)
	}
}

func (bot *Hitbot) joinChannel(channel string) {
	msgs := outMessage{Name: "message", Args: []arg{{Method: "joinChannel", Params: joinChannelParams{Channel: strings.ToLower(channel), Name: bot.name, Token: bot.auth.Token, IsAdmin: false}}}}
	var js []byte
	js, _ = json.Marshal(msgs)
	msg := "5:::" + string(js)
	bot.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (bot *Hitbot) sendMessage(channel string, text string) {
	msgs := outMessage{Name: "message", Args: []arg{{Method: "chatMsg", Params: msgParams{Channel: strings.ToLower(channel), Name: bot.name, NameColor: bot.color, Text: text}}}}
	var js []byte
	js, _ = json.Marshal(msgs)
	msg := "5:::" + string(js)
	bot.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

//MessageHandler processes messages recieved from chat server.
func (bot *Hitbot) MessageHandler() {
	for {
		_, p, err := bot.conn.ReadMessage()
		if err != nil {
			return
		}
		//log.Printf("Message: %v", string(p)) //debug info
		if string(p[:3]) == "2::" {
			bot.conn.WriteMessage(websocket.TextMessage, []byte("2::"))
			//log.Print("Ping!")
			continue
		} else if string(p[:3]) == "1::" {
			log.Print("Connection successful!")
			for _, channel := range bot.channels {
				bot.joinChannel(channel)
			}
			continue
		} else if string(p[:4]) == "5:::" {
			bot.parseMessage(p[4:])
		}
	}
}

//Connect starts connection to active server, and stores its pointer in Hitbot struct.
func (bot *Hitbot) Connect(channels ...string) {
	dialer := websocket.Dialer{}
	c, _, err := dialer.Dial("ws://"+bot.servers[bot.activeServer].ServerIP+"/socket.io/1/websocket/"+bot.connID, nil)
	if err != nil {
		log.Fatal(err)
	}
	bot.conn = c
	bot.channels = channels
}

func (bot *Hitbot) parseMessage(msg []byte) {
	var in inMessage
	if err := json.Unmarshal(msg, &in); err != nil {
		log.Fatalf("Could not parse message: %v", err)
	}
	var inArgs chatArg
	if err := json.Unmarshal([]byte(in.Args[0]), &inArgs); err != nil {
		log.Fatalf("Could not parse args: %v", err)
	}
	if inArgs.Method == "chatMsg" && inArgs.Params.Text[0] == '!' && !inArgs.Params.Buffer {
		//log.Printf("%v: %v", inArgs.Params.(map[string]interface{})["name"].(string), inArgs.Params.(map[string]interface{})["text"].(string))
		bot.dispatchCommand(inArgs.Params)
	} else if inArgs.Method == "loginMsg" {
		log.Print("Login successful!")
	}
	//log.Printf("Debug msg out: %v", string(msg))
}
