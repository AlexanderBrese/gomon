package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/configuration"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/monitoring"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/utils"
)

var (
	cfgPath string
	sigs    chan os.Signal
)

func init() {
	flag.StringVar(&cfgPath, "c", "", "config path")
	flag.Parse()
}

func main() {
	defer _recover()
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	cfg, err := parse(cfgPath)
	if err != nil {
		log.Fatal(err)
		return
	}
	path := utils.RootPath()
	cfg.Root = path
	fileChanges, err := monitoring.NewFileChanges(cfg)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		<-sigs
		fileChanges.StopWatching()
	}()

	if err := fileChanges.Watch(); err != nil {
		log.Fatal(err)
	}

}

func parse(cfgPath string) (*configuration.Configuration, error) {
	cfg, err := configuration.ParsedConfiguration(cfgPath)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}

func _recover() {
	if e := recover(); e != nil {
		log.Fatalf("PANIC: %+v", e)
	}
}
