package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	dg "github.com/bwmarrin/discordgo"
)

// Config : the structure of the config file
type Config struct {
	BotID string
}

var newMessageInstructions = map[string]func(*dg.Session, *dg.MessageCreate){
	"$headsortails": headsTails,
}

func headsTails(s *dg.Session, m *dg.MessageCreate) {
	num := rand.Intn(2)
	if num == 0 {
		s.ChannelMessageSend(m.ChannelID, "Heads")
	} else {
		s.ChannelMessageSend(m.ChannelID, "Tails")
	}
}

func help(s *dg.Session, m *dg.MessageCreate) {
	out := "The following instructions are available:\n$help\n"
	for k := range newMessageInstructions {
		out += k + "\n"
	}
	s.ChannelMessageSend(m.ChannelID, out)
}

func swear() {

}

func main() {
	var cfg Config
	runtime.GOMAXPROCS(2)
	readConfig(&cfg)

	session, err := dg.New(cfg.BotID)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer session.Close()

	if err = session.Open(); err != nil {
		fmt.Println(err)
		return
	}

	session.AddHandler(newMessage)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, syscall.SIGSEGV, syscall.SIGHUP)
	fmt.Println("Setup Complete")
	<-sc
}

func readConfig(cfg *Config) {
	raw, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(raw, cfg)
}

func newMessage(s *dg.Session, m *dg.MessageCreate) {
	if m.Content[0] == '$' {
		keyWord := strings.Split(m.Content, " ")[0]
		if val, ok := newMessageInstructions[keyWord]; ok {
			val(s, m)
		}
		if keyWord == "$help" {
			help(s, m)
		}
	}

}
