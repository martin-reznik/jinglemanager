package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"github.com/martin-reznik/jinglemanager/lib"
	"github.com/martin-reznik/jinglemanager/server"
	"github.com/martin-reznik/logger"
	"github.com/skratchdot/open-golang/open"
	"net/http"
	"os"
	"os/signal"
	"sync"
)

func main() {
	flagDoNotOpenBrowser := flag.Bool("no-browser", false, "do not open browser")
	flag.Parse()

	log := logger.NewLog(func(line *logger.LogLine) { lib.ChannelLog.Emit(lib.EventTypeLog, line) })
	log.LogSeverity[logger.DEBUG] = true

	Ctx := &lib.Context{
		Log:        log,
		Songs:      lib.NewFileList(),
		Sound:      lib.NewSoundController(log),
		Tournament: lib.NewTournament(""),
	}

	defer func() {
		lib.Save(Ctx)
		log.Close()
	}()

	lib.LoadFromFile(Ctx, "last.yml")

	httpHandler := server.HTTPHandler{Context: Ctx}
	fileHandler := server.FileProxyHandler{Context: Ctx}
	playerHandler := server.PlayerHandler{Context: Ctx}
	controlHandler := server.SoundControlHandler{Context: Ctx}
	storageHandler := server.StorageHandler{Context: Ctx}
	socketHandler := server.SocketHandler{Context: Ctx, Upgrader: &websocket.Upgrader{}}

	router := httprouter.New()
	router.GET("/", httpHandler.Index)

	router.GET("/css/*filepath", fileHandler.Static)
	router.GET("/js/*filepath", fileHandler.Static)
	router.GET("/images/*filepath", fileHandler.Static)

	router.POST("/track/add", playerHandler.Add)
	router.GET("/track/list", playerHandler.List)
	router.POST("/track/play/:id", playerHandler.Play)
	router.POST("/track/stop/:id", playerHandler.Stop)
	router.POST("/track/pause/:id", playerHandler.Pause)
	router.DELETE("/track/delete/:id", playerHandler.Delete)

	router.POST("/app/mute", controlHandler.Mute)
	router.POST("/app/unmute", controlHandler.UnMute)
	router.POST("/app/add", controlHandler.Add)
	router.DELETE("/app/delete/:id", controlHandler.Delete)
	router.GET("/app/list", controlHandler.List)

	router.GET("/save", storageHandler.Save)
	router.POST("/load", storageHandler.Load)

	router.GET("/changes", socketHandler.HandleChangeSocket)
	router.GET("/logs", socketHandler.HandleLogSocket)

	wg := sync.WaitGroup{}
	go func() {
		defer wg.Done()
		http.ListenAndServe(":8080", router)
	}()

	wg.Add(1)
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		defer wg.Done()
		for signal := range c {
			switch signal {
			case os.Interrupt:
				lib.Save(Ctx)
				return
			}
		}
	}()

	log.Info("Server is up and running, open 'http://localhost:8080' in your browser")

	if !*flagDoNotOpenBrowser {
		open.Start("http://localhost:8080")
	}
	wg.Wait()
}
