package main

import (
	"nathanman/database"
	"nathanman/discord"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

func main() {
	database.Migrate()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	quit := make(chan interface{}, 1)
	wg := new(sync.WaitGroup)
	wg.Add(1)
	go discord.Run(quit, wg)
	<-sc
	quit <- "stop"
	wg.Wait()
}
