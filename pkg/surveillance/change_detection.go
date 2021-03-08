package surveillance

import (
	"github.com/AlexanderBrese/GOATmon/pkg/configuration"
)

type ChangeDetection struct {
	environment *Environment
	control     *Control
	detection   *Detection
}

func NewChangeDetection(cfg *configuration.Configuration) (*ChangeDetection, error) {
	env, err := NewEnvironment(cfg)
	if err != nil {
		return nil, err
	}
	if err := env.Run(); err != nil {
		return nil, err
	}
	n := NewNotification()
	ctrl := NewControl(env, n)
	d, err := NewDetection(env, n)
	if err != nil {
		return nil, err
	}

	c := &ChangeDetection{
		environment: env,
		control:     ctrl,
		detection:   d,
	}

	return c, nil
}

func (c *ChangeDetection) Subscribe(sub chan bool) {
	c.detection.notification = NewSubscriberNotification(sub)
}

func (c *ChangeDetection) Start() {
	go c.detection.Run()
	c.control.Run()
	c.environment.Teardown()
}

func (c *ChangeDetection) Stop() {
	c.control.Stop()
}
