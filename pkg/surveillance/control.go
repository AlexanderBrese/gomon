package surveillance

import (
	"fmt"
)

type Control struct {
	environment   *Environment
	notification  *Notification
	stopDetecting chan bool
}

func NewControl(env *Environment, n *Notification) *Control {
	return &Control{
		environment:   env,
		notification:  n,
		stopDetecting: make(chan bool),
	}
}

func (c *Control) Stop() {
	c.stopDetecting <- true
}

func (c *Control) Run() {
	firstRun := make(chan bool, 1)
	firstRun <- true

	for {
		select {
		case <-c.stopDetecting:
			return
		case <-c.notification.ChangeDetected():
			fmt.Println("change detected")
		case <-firstRun:
			break
		}

		if c.environment.config.Reload {
			c.environment.reloader.Run()
			<-c.environment.reloader.FinishedRunning
		}
		if c.environment.config.Sync {
			c.environment.sync.Sync()
		}
	}
}
