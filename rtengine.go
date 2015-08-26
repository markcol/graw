package graw

import (
	"github.com/turnage/graw/api"
	"github.com/turnage/graw/internal/monitor"
	"github.com/turnage/graw/internal/operator"
	"github.com/turnage/redditproto"
)

// rtEngine runs bots against real time Reddit.
type rtEngine struct {
	// op is the rtEngine's operator for making reddit api calls.
	Op operator.Operator
	// monitors is a slice of the monitors rtEngine gets events from.
	Monitors []monitor.Monitor
	// Actor is the bot's interface for receiving an interface to the
	// Engine, so that it can act through its Reddit account.
	Actor api.Actor
	// Failer is the bot's interface for handling errors. rtEngine will
	// defer to this to decide what to when it encounters an error.
	Failer api.Failer
	// Loader is the bot's interface for setting up and tearing down
	// resources.
	Loader api.Loader

	// stop is a switch bots can set to signal the engine should stop.
	stop bool
}

// Reply submits a reply.
func (r *rtEngine) Reply(parentName, text string) error {
	return r.Op.Reply(parentName, text)
}

// SendMessage sends a private message.
func (r *rtEngine) SendMessage(user, subject, text string) error {
	return r.Op.Compose(user, subject, text)
}

// SelfPost makes a self (text) post to a subreddit.
func (r *rtEngine) SelfPost(subreddit, title, text string) error {
	return r.Op.Submit(subreddit, "self", title, text)
}

// LinkPost makes a link post to a subreddit.
func (r *rtEngine) LinkPost(subreddit, title, url string) error {
	return r.Op.Submit(subreddit, "link", title, url)
}

// DigestThread returns a Link with a parsed comment tree.
func (r *rtEngine) DigestThread(permalink string) (*redditproto.Link, error) {
	return r.Op.Thread(permalink)
}

// Stop is a function exposed to bots to stop the engine.
func (r *rtEngine) Stop() {
	r.stop = true
}

// Run is the main engine loop.
func (r *rtEngine) Run() error {
	if r.Loader != nil {
		r.Loader.SetUp()
		defer r.Loader.TearDown()
	}

	if r.Actor != nil {
		r.Actor.TakeEngine(r)
	}

	for !r.stop {
		for _, mon := range r.Monitors {
			if err := mon.Update(); err != nil {
				if r.fail(err) {
					return err
				}
			}
		}
	}

	return nil
}

// fail lets the bot decide whether to treat an error as a failure.
func (r *rtEngine) fail(err error) bool {
	if r.Failer == nil {
		return false
	}

	return r.Failer.Fail(err)
}
