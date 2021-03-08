package surveillance

import "github.com/AlexanderBrese/gomon/pkg/configuration"

type Gomon struct {
	environment *Environment
	control     *Refresh
	detection   *Detection
}

func NewGomon(cfg *configuration.Configuration) (*Gomon, error) {
	env, err := NewEnvironment(cfg)
	if err != nil {
		return nil, err
	}

	n := NewNotification()
	ctrl := NewRefresh(env, n)
	d, err := NewDetection(env, n)
	if err != nil {
		return nil, err
	}

	c := &Gomon{
		environment: env,
		control:     ctrl,
		detection:   d,
	}

	return c, nil
}

func (c *Gomon) Subscribe(sub chan bool) {
	c.detection.notification = NewSubscriberNotification(sub)
}

func (c *Gomon) Start() {
	go c.detection.Run()
	c.control.Run()
}

func (c *Gomon) Stop() {
	c.environment.Teardown()
}
