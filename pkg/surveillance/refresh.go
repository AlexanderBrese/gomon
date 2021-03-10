package surveillance

type Refresh struct {
	environment  *Environment
	notification *Notification
}

func NewRefresh(env *Environment, n *Notification) *Refresh {
	return &Refresh{
		environment:  env,
		notification: n,
	}
}

func (c *Refresh) Run() {
	startupRun := make(chan bool, 1)
	startupRun <- true
	for {
		select {
		case <-c.environment.stopRefreshing:
			c.notification.Stop()
			close(c.environment.stopRefreshing)
			return
		case <-c.notification.ChangeDetected():
			c.log()
		case <-startupRun:
			break
		}

		if c.reload() {
			c.sync()
		}
	}
}

func (c *Refresh) log() {
	c.environment.logger.Detection("%s", "change detected")
}

func (c *Refresh) reload() bool {
	if c.environment.config.Reload {
		c.environment.reloader.Run()
		return <-c.environment.reloader.FinishedRunning
	}
	return false
}

func (c *Refresh) sync() {
	if c.environment.config.Sync {
		c.environment.sync.Sync()
	}
}
