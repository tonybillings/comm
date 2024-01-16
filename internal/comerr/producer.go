package comerr

import "sync"

type Producer interface {
	Errors() <-chan error
}

type DefaultProducer struct {
	errorChan      chan error
	errorChanMutex sync.Mutex
}

func (p *DefaultProducer) Errors() <-chan error {
	return p.errorChan
}

func (p *DefaultProducer) ConfigureErrors(chanBufferSize int) {
	p.CloseErrors()
	p.errorChanMutex.Lock()
	p.errorChan = make(chan error, chanBufferSize)
	p.errorChanMutex.Unlock()
}

func (p *DefaultProducer) SendError(err error) {
	p.errorChanMutex.Lock()
	defer p.errorChanMutex.Unlock()

	if p.errorChan != nil {
		select {
		case p.errorChan <- err:
			return
		default:
		}
	}
}

func (p *DefaultProducer) CloseErrors() {
	p.errorChanMutex.Lock()
	defer p.errorChanMutex.Unlock()

	if p.errorChan != nil {
		select {
		case <-p.errorChan:
			p.errorChan = nil
		default:
			close(p.errorChan)
			p.errorChan = nil
		}
	}
}
