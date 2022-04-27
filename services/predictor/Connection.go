package predictor

import (
	"fmt"
	"io"
	"net"
	"sync"

	"returntypes-langserver/common/configuration"
	"returntypes-langserver/common/debug/errors"
	"returntypes-langserver/common/transfer/rpc"
)

// A connection to the predictor using the TCP protocol.
type PredictorConnection struct {
	conn           net.Conn
	connReadMutex  sync.Mutex
	connWriteMutex sync.Mutex
}

// Tries to connect to the predictor.
func (p *PredictorConnection) Connect() errors.Error {
	if p.IsConnected() {
		return errors.New(PredictorErrorTitle, "Connection is already established.")
	}
	return p.connectOverNetwork()
}

// Returns true if the predictor is (still) connected.
func (p *PredictorConnection) IsConnected() bool {
	return p.conn != nil
}

// Reads bytes from the connection with the predictor.
func (p *PredictorConnection) Read(b []byte) (int, error) {
	p.connReadMutex.Lock()
	defer p.connReadMutex.Unlock()

	if p.conn != nil {
		n, err := p.conn.Read(b)
		if err != nil {
			p.Close()
			return n, errors.Wrap(io.ErrClosedPipe, PredictorErrorTitle, "Could not read from connection")
		}
		return n, nil
	} else {
		return 0, errors.Wrap(io.ErrClosedPipe, PredictorErrorTitle, "No connection")
	}
}

// Writes bytes to the connection with the predictor.
func (p *PredictorConnection) Write(b []byte) (int, error) {
	p.connWriteMutex.Lock()
	defer p.connWriteMutex.Unlock()

	if p.conn != nil {
		n, err := p.conn.Write(b)
		if err != nil {
			p.Close()
			return n, errors.Wrap(err, PredictorErrorTitle, "Could not write to connection")
		}
		return n, nil
	} else {
		return 0, errors.Wrap(io.ErrClosedPipe, PredictorErrorTitle, "No connection p = %p", p)
	}
}

// Closes the conenction to the predictor.
func (p *PredictorConnection) Close() errors.Error {
	if p.conn != nil {
		p.conn.Close()
		p.conn = nil
	}
	return nil
}

func (p *PredictorConnection) connectOverNetwork() errors.Error {
	p.connWriteMutex.Lock()
	p.connReadMutex.Lock()
	defer p.connWriteMutex.Unlock()
	defer p.connReadMutex.Unlock()

	if p.conn != nil {
		return nil
	}

	conn, err := net.DialTimeout("tcp", p.predictorAddress(), configuration.ConnectionTimeout())
	if err == nil {
		p.conn = conn
		return nil
	} else {
		p.Close()
		var netError net.Error
		if ok := errors.As(err, &netError); ok && !netError.Timeout() && !netError.Temporary() {
			connErr := rpc.NewConnectionError(err, "An unrecoverable connection error occured", false)
			return errors.Wrap(connErr, PredictorErrorTitle, "Could not connect to the predictor")
		} else {
			connErr := rpc.NewConnectionError(err, "A connection error occured", true)
			return errors.Wrap(connErr, PredictorErrorTitle, "Could not connect to the predictor")
		}
	}
}

func (p *PredictorConnection) predictorAddress() string {
	return fmt.Sprintf("%s:%d", configuration.PredictorHost(), configuration.PredictorPort())
}

// Returns always true as the predictor connection is recoverable.
func (p *PredictorConnection) IsRecoverable() bool {
	return true
}
