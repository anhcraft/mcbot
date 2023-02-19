package main

import (
	"fmt"
	"github.com/Tnze/go-mc/bot"
	"github.com/Tnze/go-mc/bot/basic"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
)

func main() {
	max := 100
	if m, e := strconv.Atoi(os.Getenv("max")); e == nil {
		max = m
	}
	addr := os.Getenv("addr")
	if addr == "" {
		addr = "localhost"
	}
	var group sync.WaitGroup
	group.Add(max)

	var counter atomic.Int32

	for i := 0; i < max; i++ {
		go func(i int) {
			client := bot.NewClient()
			client.Auth.Name = fmt.Sprintf("bot%d", i)
			_ = basic.NewPlayer(client, basic.DefaultSettings, basic.EventsListener{
				GameStart:    nil,
				Disconnect:   nil,
				HealthChange: nil,
				Death:        nil,
				Teleported:   nil,
			})
			err := client.JoinServer(addr)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Login success: %s (%d/%d)\n", client.Auth.Name, counter.Add(1), max)
			if err = client.HandleGame(); err != nil {
				log.Fatal(err)
			}
			group.Done()
		}(i)
	}

	group.Wait()
}
