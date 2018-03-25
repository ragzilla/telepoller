// vim: tabstop=4 softtabstop=4 shiftwidth=4 noexpandtab tw=72

package telepoller

import (
	"os"

	_ "github.com/nats-io/go-nats"
	_ "github.com/ugorji/go/codec"
)

type tpPoller interface {
	Init(string) error
}

type TpFramework struct {
	poller *tpPoller
}

func NewFramework() *TpFramework {
	f := TpFramework{}
	return &f
}

func (f *TpFramework) Init(p tpPoller) error {
	f.poller = &p
	// load configuration
	config := os.Args[0]
	config += ".conf"

	// initialize poller
	if err := p.Init(config); err != nil {
		panic(err)
	}

	return nil
}

func (f *TpFramework) Run() error {
	return nil
}

func (f *TpFramework) Publish() error {
	return nil
}

func (f *TpFramework) Done() error {
	return nil
}
