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
		addr = "localhost"
	}
	password = os.Getenv("password")

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

func listenCommands() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		text := strings.TrimSpace(scanner.Text())
		args := strings.Split(text, " ")

		if len(args) == 0 || len(args[0]) == 0 || args[0] == "help" {
			log.Println("Available commands:")
			log.Println("exit")
			log.Println("join <name> [password]")
			log.Println("list")
			log.Println("quit <name>")
			log.Println("chat <name> <message>")
			log.Println("show-chat <name>")
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
