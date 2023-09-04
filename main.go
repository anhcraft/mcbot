package main

import (
	"bufio"
	"github.com/Tnze/go-mc/bot"
	"log"
	"mcbot/handler"
	"os"
	"strings"
	"sync"
)

var addr string
var password string
var group sync.WaitGroup
var players = make(map[string]*handler.FakePlayer)

func main() {
	addr = os.Getenv("addr")
	if addr == "" {
		addr = "minehay.com"
	}
	password = os.Getenv("password")
	if password == "" {
		password = "minehay32123"
	}

	group.Add(1)
	go listenCommands()

	group.Wait()
}

func addBot(name string, password string) {
	group.Add(1)

	go func() {
		log.Printf("Creating player %s...", name)
		fakePlayer := &handler.FakePlayer{
			Client:   bot.NewClient(),
			Password: password,
		}
		players[name] = fakePlayer
		fakePlayer.Join(addr, name)
		group.Done()
	}()
}

func printCmd() {
	log.Println("Available commands:")
	log.Println("list: list all bots")
	log.Println("join <name> [password]: join a new bot")
	log.Println("quit <name>: quit a bot")
	log.Println("chat <name> <message>: let a bot chat something (or /command)")
	log.Println("show-chat <name>: show chat received from a bot")
	log.Println("exit: shutdown")
}

func listenCommands() {
	printCmd()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		args := strings.Split(text, " ")

		if len(args) == 0 || len(args[0]) == 0 || args[0] == "help" {
			printCmd()
			continue
		}

		if args[0] == "exit" {
			os.Exit(1)
		} else if args[0] == "join" {
			if len(args) == 2 {
				addBot(args[1], password)
			} else if len(args) == 3 {
				addBot(args[1], args[2])
			} else {
				log.Println("Usage: join <name> [password]")
			}
		} else if args[0] == "list" {
			log.Println("All players:")
			for name, _ := range players {
				log.Print(name, " ")
			}
		} else if args[0] == "quit" {
			if len(args) == 2 {
				if player, ok := players[args[1]]; ok {
					player.Quit()
				} else {
					log.Println("Player not found")
				}
			} else {
				log.Println("Usage: quit <name>")
			}
		} else if args[0] == "chat" {
			if len(args) >= 3 {
				if player, ok := players[args[1]]; ok {
					player.Chat(strings.Join(args[2:], " "))
				} else {
					log.Println("Player not found")
				}
			} else {
				log.Println("Usage: chat <name> <message>")
			}
		} else if args[0] == "show-chat" {
			if len(args) >= 1 {
				if player, ok := players[args[1]]; ok {
					player.ShowChat = !player.ShowChat
					log.Printf("Show chat is now %t", player.ShowChat)
				} else {
					log.Println("Player not found")
				}
			} else {
				log.Println("Usage: show-chat <name>")
			}
		}
	}
}
