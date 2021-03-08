package surveillance

type Notification struct {
	subscription chan bool
	change       chan bool
}

const changes = 1000

func NewNotification() *Notification {
	return NewSubscriberNotification(nil)
}

func NewSubscriberNotification(sub chan bool) *Notification {
	return &Notification{
		subscription: sub,
		change:       make(chan bool, changes),
	}
}

func (n *Notification) Stop() {
	close(n.change)
	if n.subscription != nil {
		close(n.subscription)
	}
}

func (n *Notification) NotfiyChange() {
	n.change <- true
	if n.subscription != nil {
		n.subscription <- true
	}
}

func (n *Notification) NotifyNoChange() {
	if n.subscription != nil {
		n.subscription <- false
	}
}

func (n *Notification) ChangeDetected() chan bool {
	return n.change
}
