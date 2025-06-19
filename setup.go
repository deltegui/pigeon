package pigeon

import (
	"context"
	"time"

	wmail "github.com/wneessen/go-mail"
)

type MailDestination struct {
	To  []string
	CC  []string
	BCC []string
}

type MailSenderConfiguration struct {
	Server          string
	From            string
	User            string
	Password        string
	TimeoutSeconds  int
	AsyncBufferSize int
	BaseSystemUrl   string
}

type MailApi struct {
	config MailSenderConfiguration
}

func (sender MailApi) GetConfiguration() MailSenderConfiguration {
	return sender.config
}

func (sender MailApi) PrepareMessage(destination MailDestination) (*wmail.Msg, error) {
	msg := wmail.NewMsg()
	if err := msg.From(sender.config.From); err != nil {
		return nil, err
	}
	if err := msg.To(destination.To...); err != nil {
		return nil, err
	}

	for _, cc := range destination.CC {
		if err := msg.AddCc(cc); err != nil {
			return nil, err
		}
	}
	for _, bcc := range destination.BCC {
		if err := msg.AddBcc(bcc); err != nil {
			return nil, err
		}
	}

	return msg, nil
}

func (sender MailApi) SendWithContext(ctx context.Context, msg *wmail.Msg) error {
	client, err := wmail.NewClient(
		sender.config.Server,
		wmail.WithTLSPortPolicy(wmail.TLSMandatory),
		wmail.WithSMTPAuth(wmail.SMTPAuthPlain),
		wmail.WithUsername(sender.config.User),
		wmail.WithPassword(sender.config.Password))
	if err != nil {
		return err
	}
	if err = client.DialAndSendWithContext(ctx, msg); err != nil {
		return err
	}
	return nil
}

func (sender MailApi) Send(msg *wmail.Msg) error {
	ctx, done := context.WithTimeout(context.Background(), 1*time.Minute)
	defer done()
	return sender.SendWithContext(ctx, msg)
}
