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

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
}

func main() {
	defer _recover()
	_onInterrupt()

	cfg, err := parse(cfgPath)
	if err != nil {
		log.Fatal(err)
		return
	}

	_, err = monitoring.NewFileChanges(cfg)
	if err != nil {
		log.Fatal(err)
		return
	}

	// Run
}

func parse(cfgPath string) (*configuration.Configuration, error) {
	absPath, err := utils.AbsolutePath(cfgPath)
	if err != nil {
		return nil, err
	}
	cfg, err := configuration.ParsedConfiguration(absPath)
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

func _onInterrupt() {
	<-sigs
	// Stop
}
