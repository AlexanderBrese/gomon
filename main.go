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

var cfgPath string

func init() {
	flag.StringVar(&cfgPath, "c", "", "relative config path")
	flag.Parse()
}

func main() {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	defer _recover()

	cfg, err := parseConfig(cfgPath)
	if err != nil {
		log.Fatal(err)
	}
	cd, err := changeDetection(cfg)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		<-sigs
		//<-time.After(2 * time.Second)
		cd.Stop()
	}()

	cd.Start()
}

func _recover() {
	if e := recover(); e != nil {
		log.Fatalf("PANIC: %+v", e)
	}
}

func changeDetection(cfg *configuration.Configuration) (*surveillance.Gomon, error) {
	changeDetection, err := surveillance.NewGomon(cfg)
	if err != nil {
		return nil, err
	}
	return changeDetection, nil
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
