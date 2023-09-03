package handler

import (
	"errors"
	"fmt"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"github.com/Tnze/go-mc/bot/msg"
	"github.com/Tnze/go-mc/bot/playerlist"
	"github.com/Tnze/go-mc/chat"
	"github.com/Tnze/go-mc/data/packetid"
	"github.com/Tnze/go-mc/net/packet"
	"log"
	"strings"
	"time"
)

type FakePlayer struct {
	Client           *bot.Client
	playerList       *playerlist.PlayerList
	player           *basic.Player
	msgManager       *msg.Manager
	Password         string
	islandChecked    bool
	ShowChat         bool
	CurrentContainer packet.VarInt
	ScreenState      packet.VarInt
	Cursor           Slot
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
	p.Client.Events.AddListener(bot.PacketHandler{
		Priority: 64, ID: packetid.ClientboundOpenScreen,
		F: p.onOpenScreen,
	})
	p.Client.Events.AddListener(bot.PacketHandler{
		Priority: 64, ID: packetid.ClientboundContainerSetContent,
		F: p.onSetScreenContent,
	})
	p.Client.Events.AddListener(bot.PacketHandler{
		Priority: 64, ID: packetid.ClientboundContainerSetSlot,
		F: p.onSetScreenSlot,
	})
	p.Client.Events.AddListener(bot.PacketHandler{
		Priority: 64, ID: packetid.ClientboundContainerClose,
		F: p.onCloseScreen,
	})

	err := p.Client.JoinServer(addr)
	if err != nil {
		log.Printf("[%s] %v\n", p.getName(), err)
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
			log.Printf("[%s] %v\n", p.getName(), err)
			break
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
		time.Sleep(time.Duration(2) * time.Second)
		p.visitServer()
	}()
}

func (p FakePlayer) visitServer() {
	go func() {
		log.Printf("[%s] Trying to visit server....\n", p.getName())
		p.Chat("/menu")
	}()
}

func (p FakePlayer) Quit() {
	err := p.Client.Close()
	if err != nil {
		log.Printf("[%s] %v\n", p.getName(), err)
	}
}

func (p FakePlayer) onOpenScreen(pk packet.Packet) error {
	var (
		ContainerID packet.VarInt
		Type        packet.VarInt
		Title       chat.Message
	)
	if err := pk.Scan(&ContainerID, &Type, &Title); err != nil {
		log.Printf("[%s] %v\n", p.getName(), err)
		return err
	} else {
		log.Printf("[%s] Opened Container: %d, Type: %d, Title: %s\n", p.getName(), ContainerID, Type, Title)
		p.CurrentContainer = ContainerID
		if strings.Contains(Title.String(), "sᴇʀvᴇʀ sᴇʟᴇcтoʀ") {
			p.clickContainer(10, 0, 0, ChangedSlots{}, &p.Cursor)
		}
	}
	return nil
}

func (p FakePlayer) onSetScreenContent(pk packet.Packet) error {
	var (
		ContainerID packet.UnsignedByte
		StateID     packet.VarInt
		SlotData    []Slot
		CarriedItem Slot
	)
	if err := pk.Scan(
		&ContainerID,
		&StateID,
		packet.Array(&SlotData),
		&CarriedItem,
	); err != nil {
		log.Printf("[%s] %v\n", p.getName(), err)
		return err
	} else {
		p.ScreenState = StateID
	}
	return nil
}

func (p FakePlayer) onSetScreenSlot(pk packet.Packet) error {
	var (
		ContainerID packet.Byte
		StateID     packet.VarInt
		SlotID      packet.Short
		SlotData    Slot
	)
	if err := pk.Scan(&ContainerID, &StateID, &SlotID, &SlotData); err != nil {
		log.Printf("[%s] %v\n", p.getName(), err)
		return err
	}

	p.ScreenState = StateID

	if ContainerID == -1 && SlotID == -1 {
		p.Cursor = SlotData
	}

	return nil
}

func (p FakePlayer) onCloseScreen(pk packet.Packet) error {
	var ContainerID packet.UnsignedByte
	if err := pk.Scan(&ContainerID); err != nil {
		log.Printf("[%s] %v\n", p.getName(), err)
		return err
	} else if int(p.CurrentContainer) == int(ContainerID) {
		log.Printf("[%s] Closed Container: %d\n", p.getName(), ContainerID)
	}
	return nil
}

func (p FakePlayer) clickContainer(slot int16, button byte, mode int32, slots ChangedSlots, carried *Slot) error {
	return p.Client.Conn.WritePacket(packet.Marshal(
		packetid.ServerboundContainerClick,
		packet.UnsignedByte(p.CurrentContainer),
		p.ScreenState,
		packet.Short(slot),
		packet.Byte(button),
		packet.VarInt(mode),
		slots,
		carried,
	))
}
