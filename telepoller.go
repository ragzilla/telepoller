// vim: tabstop=4 softtabstop=4 shiftwidth=4 noexpandtab tw=72

package telepoller

import (
	"bytes"
	"fmt"
	"os"
	"os/signal"
	"path"
	"sync"
	"syscall"

	client "github.com/influxdata/influxdb/client/v2"
	nats "github.com/nats-io/go-nats"
)

type tpPoller interface {
	Init(*TpFramework, string) error
	NewJob(*TpJob, func())
}

type TpFramework struct {
	poller tpPoller
	nc     *nats.Conn
	ec     *nats.EncodedConn

	inqueue  string
	outqueue string

	outbuffer bytes.Buffer

	quit chan os.Signal

	jobs chan *TpJob
}

type TpJob struct {
	Hosts  map[string]string `json:"hosts"`
	Params map[string]string `json:"params"`
}

func NewFramework() *TpFramework {
	f := TpFramework{}
	return &f
}

func (f *TpFramework) Init(p tpPoller) error {
	f.poller = p
	// load configuration
	config := path.Base(os.Args[0])
	f.inqueue = config + ".request"
	f.outqueue = config + ".response"
	config += ".conf"

	// initialize poller
	if err := p.Init(f, config); err != nil {
		return err
	}

	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return err
	}
	f.nc = nc
	ec, err := nats.NewEncodedConn(f.nc, nats.JSON_ENCODER)
	if err != nil {
		return err
	}
	f.ec = ec

	f.jobs = make(chan *TpJob, 1)
	f.quit = make(chan os.Signal, 1)
	signal.Notify(f.quit, os.Interrupt, syscall.SIGTERM)

	return nil
}

func (f *TpFramework) Run() error {
	wg := &sync.WaitGroup{}
	fmt.Println("entering main loop")

	sub, err := f.ec.BindRecvQueueChan(f.inqueue, f.inqueue, f.jobs)
	if err != nil {
		return err
	}
	defer sub.Unsubscribe()

	for {
		select {
		case j := <-f.jobs:
			if len(j.Hosts) == 0 || len(j.Params) == 0 {
				continue
			}
			wg.Add(1)
			f.poller.NewJob(j, func() {
				// fire wg.Done() in NewJob callback
				wg.Done()
			})
		case _ = <-f.quit:
			// gather outstanding requests
			sub.Unsubscribe()
			wg.Wait()
			fmt.Println("exiting main loop")
			return nil
		}
	}
}

func (f *TpFramework) Publish(pt *client.Point) error {
	f.outbuffer.Reset()
	f.outbuffer.WriteString("foobar")
	return f.nc.Publish(f.outqueue, f.outbuffer.Bytes())
}

func (f *TpFramework) Done() error {
	if f.ec != nil {
		f.ec.Close()
	}
	if f.nc != nil {
		f.nc.Close()
	}
	return nil
}
