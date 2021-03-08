package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexanderBrese/Gomon/pkg/configuration"
	"github.com/AlexanderBrese/Gomon/pkg/surveillance"
	"github.com/AlexanderBrese/Gomon/pkg/utils"
)

var cfgPath string

func init() {
	flag.StringVar(&cfgPath, "c", "", "relative config path")
	flag.Parse()
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer _recover()

	cfg, err := parse(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	gomon, err := surveillance.NewGomon(cfg)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-sigs
		gomon.Stop()
	}()

	gomon.Start()
}

func _recover() {
	if e := recover(); e != nil {
		log.Fatalf("PANIC: %+v", e)
	}
}

func parse(cfgPath string) (*configuration.Configuration, error) {
	absPath := ""
	if cfgPath != "" {
		var err error
		absPath, err = utils.CurrentAbsolutePath(cfgPath)
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
