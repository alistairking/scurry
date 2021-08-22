package scurry

import (
	"bufio"
	"context"
	"net"
	"strconv"
	"strings"
	"sync"
)

const (
	CMD_Q_LEN = 100
)

// Simple wrapper around a TCP connection "attached" to a scamper daemon
type ScAttach struct {
	log    Logger
	moreCh chan struct{} // number of unused "MORE"s from scamper

	txWorkerCancel context.CancelFunc
	txWorkerWg     *sync.WaitGroup

	rxWorkerCancel context.CancelFunc
	rxWorkerWg     *sync.WaitGroup

	cmdQ  chan string // queue of commands to send to scamper
	dataQ chan string // queue of data responses received from scamper
	errQ  chan string // queue of error responses received from scamper

	conn  net.Conn
	txBuf *bufio.Writer
}

func NewScAttach(log Logger, url string) (*ScAttach, error) {
	txWorkerCtx, txWorkerCancel := context.WithCancel(context.Background())
	rxWorkerCtx, rxWorkerCancel := context.WithCancel(context.Background())

	a := &ScAttach{
		log:    initLogger(log, "sc-attach"),
		moreCh: make(chan struct{}, CMD_Q_LEN),

		txWorkerCancel: txWorkerCancel,
		txWorkerWg:     &sync.WaitGroup{},
		rxWorkerCancel: rxWorkerCancel,
		rxWorkerWg:     &sync.WaitGroup{},

		cmdQ:  make(chan string, CMD_Q_LEN),
		dataQ: make(chan string, CMD_Q_LEN),
		errQ:  make(chan string, CMD_Q_LEN),
	}

	// connect to scamper
	if err := a.initConnection(url); err != nil {
		return nil, err
	}

	// boot up our workers
	a.rxWorkerWg.Add(1)
	go a.rxWorker(rxWorkerCtx)
	a.txWorkerWg.Add(1)
	go a.txWorker(txWorkerCtx)

	return a, nil
}

func (a *ScAttach) CommandQueue() chan string {
	return a.cmdQ
}

func (a *ScAttach) ResultQueue() chan string {
	return a.dataQ
}

func (a *ScAttach) ErrorQueue() chan string {
	return a.errQ
}

func (a *ScAttach) Close() {
	// cancel our tx worker and wait for it to be done
	a.txWorkerCancel()
	a.txWorkerWg.Wait()
	// and do the same for our rx worker
	a.rxWorkerCancel()
	a.rxWorkerWg.Wait()
	// and now shut down our connection to scamper
	a.conn.Close()
}

// private methods

func (a *ScAttach) initConnection(url string) error {
	// TODO: better unix socket detection
	var conn net.Conn
	var err error
	if strings.Contains(":", url) {
		conn, err = net.Dial("tcp", url)
	} else {
		conn, err = net.Dial("unix", url)
	}
	if err != nil {
		return err
	}
	a.conn = conn

	// create buffer for tx
	a.txBuf = bufio.NewWriter(a.conn)
	// send our attach command
	a.sendCmd("attach format json")

	return nil
}

func (a *ScAttach) sendCmd(cmd string) {
	// this assumes that there is MORE available
	a.log.Debug().
		Str("command", cmd).
		Msgf("Sending command to scamper")
	a.txBuf.WriteString(cmd)
	a.txBuf.WriteString("\n")
	a.txBuf.Flush()
}

func (a *ScAttach) handleResponse(resp string) {
	if strings.HasPrefix(resp, "OK") {
		// ignore these, they're expected
		a.log.Debug().
			Msgf("Got OK from scamper")
		return
	}

	if resp == "MORE" {
		select {
		case a.moreCh <- struct{}{}:
			// cool
			a.log.Debug().
				Int("mores", len(a.moreCh)).
				Msgf("Got MORE from scamper")
		default:
			// not cool
			a.log.Warn().Msgf("More chan full, dropping MORE")
		}
		return
	}

	if strings.HasPrefix(resp, "DATA") {
		// TODO: could cross-check rx data against this
		d := strings.Split(resp, " ")
		dLen, _ := strconv.Atoi(d[1])
		a.log.Debug().
			Int("data-len", dLen).
			Msgf("Scamper data incoming")
		return
	}

	if strings.HasPrefix(resp, "ERR") {
		a.errQ <- resp[4:]
		return
	}

	// otherwise, this must be a result, fire it off
	a.dataQ <- resp
}

func (a *ScAttach) rxWorker(ctx context.Context) {
	defer func() {
		a.rxWorkerWg.Done() // signal to close that we're done
		close(a.dataQ)
		close(a.errQ)
	}()

	// start up a goroutine to chunk rx into lines
	rxChan := make(chan string, CMD_Q_LEN)
	go a.scamperRx(rxChan)

	for {
		select {
		case resp := <-rxChan:
			a.handleResponse(resp)

		case <-ctx.Done():
			// canceled, just give up
			return
		}
	}
}

func (a *ScAttach) scamperRx(outCh chan string) {
	rxBuf := bufio.NewReader(a.conn)
	scanner := bufio.NewScanner(rxBuf)
	for scanner.Scan() {
		outCh <- scanner.Text()
	}
	close(outCh)
	a.log.Info().Msgf("Scamper rx loop ending")
}

func (a *ScAttach) txWorker(ctx context.Context) {
	defer func() {
		close(a.cmdQ)
		a.txWorkerWg.Done()
	}()

	for {
		select {
		case cmd := <-a.cmdQ:
			// we have a measurement, wait until scamper wants it
			<-a.moreCh
			// alright, good to go, fire it off
			a.sendCmd(cmd)

		case <-ctx.Done():
			a.log.Info().Msgf("TX worker shutting down")
			return
		}
	}
}
