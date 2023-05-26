package handler

import (
	"errors"
	"fmt"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"github.com/Tnze/go-mc/bot/msg"
	"github.com/Tnze/go-mc/bot/playerlist"
	"github.com/Tnze/go-mc/chat"
	"log"
	"strings"
	"time"
)

type FakePlayer struct {
	Client        *bot.Client
	playerList    *playerlist.PlayerList
	player        *basic.Player
	msgManager    *msg.Manager
	Password      string
	islandChecked bool
	ShowChat      bool
}

func (p FakePlayer) getName() string {
	return p.Client.Auth.Name
}

func (p FakePlayer) Join(addr string, name string) {
	p.Client.Auth.Name = name
	p.player = basic.NewPlayer(p.Client, basic.Settings{
		Locale:              "en_US", // ^_^
		ViewDistance:        7,
		ChatMode:            0,
		ChatColors:          true,
		DisplayedSkinParts:  basic.Jacket | basic.LeftSleeve | basic.RightSleeve | basic.LeftPantsLeg | basic.RightPantsLeg | basic.Hat,
		MainHand:            1,
		EnableTextFiltering: false,
		AllowListing:        true,
		Brand:               "vanilla",
	}, basic.EventsListener{
		GameStart: func() error {
			return p.onGameStart()
		},
		Disconnect: func(reason chat.Message) error {
			return p.onDisconnect(reason)
		},
		HealthChange: nil,
		Death: func() error {
			return p.onDeath()
		},
		Teleported: nil,
	})

	err := p.Client.JoinServer(addr)
	if err != nil {
		log.Fatalf("[%s] %v\n", p.getName(), err)
	}

	p.playerList = playerlist.New(p.Client)
	p.msgManager = msg.New(p.Client, p.player, p.playerList, msg.EventsHandler{
		SystemChat: func(msg chat.Message, overlay bool) error {
			if !overlay && p.ShowChat {
				p.onChat(msg)
			}
			return nil
		},
		PlayerChatMessage: func(msg chat.Message, validated bool) error {
			if p.ShowChat {
				p.onChat(msg)
			}
			return nil
		},
		DisguisedChat: func(msg chat.Message) error {
			if p.ShowChat {
				p.onChat(msg)
			}
			return nil
		},
	})

	var perr bot.PacketHandlerError
	for {
		if err = p.Client.HandleGame(); err == nil {
			panic(fmt.Sprintf("[%s] HandleGame never return nil\n", p.getName()))
		}
		if errors.As(err, &perr) {
			log.Printf("[%s] %v\n", p.getName(), perr)
		} else {
			log.Fatalf("[%s] %v\n", p.getName(), err)
		}
	}
}

func (p FakePlayer) onGameStart() error {
	log.Printf("[%s] Game Started\n", p.getName())
	p.runLoginTask()
	return nil
}

func (p FakePlayer) onDisconnect(reason chat.Message) error {
	log.Printf("[%s] Disconnected: %s\n", p.getName(), reason)
	return nil
}

func (p FakePlayer) onDeath() error {
	log.Printf("[%s] Died\n", p.getName())
	go func() {
		log.Printf("[%s] Trying to respawn...\n", p.getName())
		time.Sleep(time.Second * 5)
		if err := p.player.Respawn(); err != nil {
			log.Printf("[%s] %v\n", p.getName(), err)
		}
	}()
	return nil
}

func (p FakePlayer) onChat(message chat.Message) {
	log.Printf("[%s] Msg: %s\n", p.getName(), message)

	if !p.islandChecked && strings.Contains(message.String(), "Bạn không có một hòn đảo.") {
		p.islandChecked = true
		log.Printf("[%s] Trying to create new island...\n", p.getName())
		p.Chat("/is create " + p.getName() + " mapskyblock-new")
	}
}

func (p FakePlayer) Chat(message string) {
	if err := p.msgManager.SendMessage(message); err != nil {
		log.Printf("[%s] %v\n", p.getName(), err)
	}
}

func (p FakePlayer) runLoginTask() {
	go func() {
		time.Sleep(time.Duration(3) * time.Second)
		log.Printf("[%s] Logging in....\n", p.getName())
		p.Chat("/login " + p.Password)
		p.runIsTask()
	}()
}

func (p FakePlayer) runIsTask() {
	go func() {
		time.Sleep(time.Duration(3) * time.Second)
		log.Printf("[%s] Trying to go to island....\n", p.getName())
		p.Chat("/is go")
	}()
}

func (p FakePlayer) Quit() {
	err := p.Client.Close()
	if err != nil {
		log.Printf("[%s] %v\n", p.getName(), err)
	}
}
