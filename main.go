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

type config struct {
	BotID string
}

type memory struct {
	ChristianChannels []string
}

type swearWord struct {
	word     string
	language string
}

var newMessageInstructions = map[string]func(*dg.Session, *dg.MessageCreate){
	"$headsortails": headsTails,
	"$christian":    christian,
}

func christian(s *dg.Session, m *dg.MessageCreate) {
	for i, v := range mem.ChristianChannels {
		if v == m.ChannelID {
			mem.ChristianChannels = append(mem.ChristianChannels[:i], mem.ChristianChannels[i+1:]...)
			s.ChannelMessageSend(m.ChannelID, "This is no longer a christian channel")
			return
		}
	}
	mem.ChristianChannels = append(mem.ChristianChannels, m.ChannelID)
	s.ChannelMessageSend(m.ChannelID, "This is now a christian channel")
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

func swear(s *dg.Session, m *dg.MessageCreate) {
	if _, ok := swears[m.Author.ID+m.ChannelID]; !ok {
		swears[m.Author.ID+m.ChannelID] = 0
	}
	index := swears[m.Author.ID+m.ChannelID]
	if index >= len(anger) {
		index = len(anger) - 1
	}
	phrase := anger[index]
	oops := false
	if strings.Contains(phrase, "%s") {
		phrase = fmt.Sprintf(phrase, m.Author.Username)
	}
	if strings.Contains(phrase, "<oops>") {
		phrase = strings.Replace(phrase, "<oops>", "", -1)
		oops = true
	}
	s.ChannelMessageSend(m.ChannelID, phrase)
	if oops {
		s.ChannelMessageSend(m.ChannelID, "oops")
	}
	swears[m.Author.ID+m.ChannelID]++
}

var cfg config
var mem memory
var swears map[string]int
var swearWords []map[string]string
var anger []string

func main() {
	runtime.GOMAXPROCS(2)
	readConfig(&cfg, "config.json")
	readConfig(&mem, "memory.json")
	readConfig(&swears, "swears.json")
	readConfig(&swearWords, "DirtyWords.json")
	readConfig(&anger, "anger.json")
	if swears == nil {
		swears = make(map[string]int)
	}

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
	writeConfig(&mem, "memory.json")
	writeConfig(&swears, "swears.json")
}

func readConfig(cfg interface{}, fileName string) {
	raw, err := ioutil.ReadFile(fileName)
	if err != nil {
		fmt.Println(err)
	}
	json.Unmarshal(raw, cfg)
}

func writeConfig(cfg interface{}, fileName string) {
	b, err := json.Marshal(cfg)
	if err != nil {
		fmt.Println(err)
	}
	ioutil.WriteFile(fileName, b, os.FileMode(0))
}

func newMessage(s *dg.Session, m *dg.MessageCreate) {
	if !m.Author.Bot {
		if m.Content[0] == '$' {
			keyWord := strings.Split(m.Content, " ")[0]
			if val, ok := newMessageInstructions[keyWord]; ok {
				val(s, m)
			}
			if keyWord == "$help" {
				help(s, m)
			}
		}

		for _, word := range strings.Split(m.Content, " ") {
			found := false
			for _, swearWord := range swearWords {
				if swearWord["word"] == strings.ToLower(word) {
					for _, channelid := range mem.ChristianChannels {
						if channelid == m.ChannelID {
							swear(s, m)
							break
						}
					}
					found = true
					break
				}
			}
			if found {
				break
			}
		}
	}
}
