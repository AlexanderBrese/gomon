package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/AlexanderBrese/GOATmon/pkg/configuration"
	"github.com/AlexanderBrese/GOATmon/pkg/surveillance"
	"github.com/AlexanderBrese/GOATmon/pkg/utils"
)

var (
	sigs            chan os.Signal
	changeDetection *surveillance.ChangeDetection
)

func init() {
	cfgPath := parseInput()
	cfg, err := parseConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	initChangeDetection(cfg)
}

func parseInput() string {
	var cfgPath string
	flag.StringVar(&cfgPath, "c", "", "relative config path")
	flag.Parse()
	return cfgPath
}

func initChangeDetection(cfg *configuration.Configuration) {
	var err error
	changeDetection, err = surveillance.NewChangeDetection(cfg)
	if err != nil {
		log.Fatal(err)
	}
}

func parseConfig(cfgPath string) (*configuration.Configuration, error) {
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

func main() {
	defer _recover()

	prepareExit()
	go onExit()

	run()
}

func _recover() {
	if e := recover(); e != nil {
		log.Fatalf("PANIC: %+v", e)
	}
}

func prepareExit() {
	sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
}

func onExit() {
	<-sigs
	changeDetection.Stop()
}

func run() {
	changeDetection.Start()
}
