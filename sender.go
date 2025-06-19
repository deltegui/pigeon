package pigeon

import (
	"fmt"
	"log"
)

type Type string

type MailHandler func(api MailApi, req SendRequest) error

type SendRequest struct {
	MailType Type
	Model    any
	Setup    MailDestination
}

type Sender interface {
	Send(req SendRequest) error
	RegisterMailType(t Type, handler MailHandler)
}

type AsyncSender struct {
	config   MailSenderConfiguration
	handlers map[Type]MailHandler
	Input    chan SendRequest
}

const defaultBuffer int = 10

func NewAsyncMailSender(config MailSenderConfiguration) *AsyncSender {
	bufferSize := defaultBuffer

	if config.AsyncBufferSize > 0 {
		bufferSize = config.AsyncBufferSize
	}
	c := make(chan SendRequest, bufferSize)

	return &AsyncSender{
		config:   config,
		handlers: map[Type]MailHandler{},
		Input:    c,
	}
}

func (sender *AsyncSender) Start() {
	go func() {
		log.Println("Mailer async thread started to listen requests...")
		api := MailApi{
			config: sender.config,
		}
		for {
			req := <-sender.Input
			log.Println("Waiting mail request to arrive...")
			impl, ok := sender.handlers[req.MailType]
			log.Println("Send mail for type", req.MailType)
			if !ok {
				err := fmt.Errorf("AyncMailSender: cannot find mail type: '%s'", req.MailType)
				log.Println(err)
			}
			if err := impl(api, req); err != nil {
				log.Println(err)
			}
			log.Println("Mail sended")
		}
	}()
}

func (sender *AsyncSender) Send(req SendRequest) error {
	sender.Input <- req
	return nil
}

func (sender *AsyncSender) RegisterMailType(t Type, handler MailHandler) {
	sender.handlers[t] = handler
}

type SyncSender struct {
	config   MailSenderConfiguration
	handlers map[Type]MailHandler
}

func NewSyncMailSender(config MailSenderConfiguration) *SyncSender {
	return &SyncSender{
		config:   config,
		handlers: map[Type]MailHandler{},
	}
}

func (sync SyncSender) Send(req SendRequest) error {
	api := MailApi{
		config: sync.config,
	}
	impl, ok := sync.handlers[req.MailType]
	if !ok {
		err := fmt.Errorf("SyncMailSender: cannot find mail type: '%s'", req.MailType)
		log.Println(err)
		return err
	}
	if err := impl(api, req); err != nil {
		log.Println(err)
		return err
	}
	return nil
}

func (sender *SyncSender) RegisterMailType(t Type, handler MailHandler) {
	sender.handlers[t] = handler
}
