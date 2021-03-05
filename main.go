package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexanderBrese/go-server-browser-reload/pkg/configuration"
	"github.com/AlexanderBrese/go-server-browser-reload/pkg/surveillance"
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

	cfg, err := parse("test.toml")
	if err != nil {
		log.Fatal(err)
		return
	}
	fileChanges, err := surveillance.NewFileChanges(cfg)
	if err != nil {
		log.Fatal(err)
	}
	if err := fileChanges.Init(); err != nil {
		log.Fatal(err)
	}
	go func() {
		<-sigs
		fileChanges.StopWatching()
	}()

	if err := fileChanges.Surveil(); err != nil {
		log.Fatal(err)
	}

}

func parse(cfgPath string) (*configuration.Configuration, error) {
	absPath := ""
	if cfgPath != "" {
		var err error
		absPath, err = utils.AbsolutePath(cfgPath)
		if err != nil {
			return nil, err
		}
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
