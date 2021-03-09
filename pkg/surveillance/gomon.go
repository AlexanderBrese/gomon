package surveillance

import "github.com/AlexanderBrese/gomon/pkg/configuration"

type Gomon struct {
	environment *Environment
	control     *Refresh
	detection   *Detection
}

func NewGomon(cfg *configuration.Configuration) *Gomon {
	env, err := NewEnvironment(cfg)
	if err != nil {
		env.logger.Main("error: during environment initialization: %s", err)
		return nil
	}

	n := NewNotification()
	ctrl := NewRefresh(env, n)
	d, err := NewDetection(env, n)
	if err != nil {
		env.logger.Main("error: during detection initialization: %s", err)
	}

	c := &Gomon{
		environment: env,
		control:     ctrl,
		detection:   d,
	}

	return c
}

func (c *Gomon) Subscribe(sub chan bool) {
	c.detection.notification = NewSubscriberNotification(sub)
}

func (c *Gomon) Start() {
	go func() {
		if err := c.detection.Run(); err != nil {
			c.environment.logger.Main("error: during detection: %s", err)
			return
		}
	}()
	c.control.Run()
}

func (c *Gomon) Stop() {
	if err := c.environment.Teardown(); err != nil {
		c.environment.logger.Main("error: during environment teardown: %s", err)
		return
	}
}
