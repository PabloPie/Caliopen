// Copyleft (ɔ) 2017 The Caliopen contributors.
// Use of this source code is governed by a GNU AFFERO GENERAL PUBLIC
// license (AGPL) that can be found in the LICENSE file.

package caliopen_smtp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	broker "github.com/CaliOpen/Caliopen/src/backend/brokers/go.emails"
	. "github.com/CaliOpen/Caliopen/src/backend/defs/go-objects"
	log "github.com/Sirupsen/logrus"
	"gopkg.in/gomail.v2"
	"io"
	"net/smtp"
	"strconv"
	"strings"
	"sync"
	"time"
)

type submitter struct {
	config          LDAConfig
	workersCountMux sync.Mutex
	runningWorkers  int
	submitChan      chan *broker.SmtpEmail
}

type smtpSender struct {
	smtpClient
	d *gomail.Dialer
}

type smtpClient interface {
	Hello(string) error
	Extension(string) (bool, string)
	StartTLS(*tls.Config) error
	Auth(smtp.Auth) error
	Mail(string) error
	Rcpt(string) error
	Data() (io.WriteCloser, error)
	Quit() error
	Close() error
}

func (lda *Lda) newSubmitter() (submit *submitter, err error) {

	submit = &submitter{}
	submit.config = lda.Config.LDAConfig
	submit.submitChan = make(chan *broker.SmtpEmail)

	return
}

func (lda *Lda) runSubmitterAgent() {

	for email := range lda.brokerConnectors.Egress {
		go func(email *broker.SmtpEmail) {
			lda.outboundListener.workersCountMux.Lock()
			if lda.outboundListener.runningWorkers < lda.Config.AppConfig.OutWorkers {
				go lda.OutboundWorker()
				lda.outboundListener.runningWorkers++
			}
			lda.outboundListener.workersCountMux.Unlock()

			//submit email
			lda.outboundListener.submitChan <- email
		}(email)
	}

}

/*  OutboundWorker dials to static MTA and maintains connection open to handle outbound deliveries,
then close the connection if no email comes in for 30 sec.
should be launched in a goroutine
*/
func (lda *Lda) OutboundWorker() {
	c := lda.Config.AppConfig
	d := gomail.NewDialer(c.SubmitAddress, c.SubmitPort, c.SubmitUser, c.SubmitPassword)
	defer func() {
		lda.outboundListener.workersCountMux.Lock()
		lda.outboundListener.runningWorkers--
		lda.outboundListener.workersCountMux.Unlock()
	}()

	var smtp_sender gomail.SendCloser
	var smtp_remote_sender gomail.SendCloser
	var err error
	open := false
	for {
		select {
		case outcoming, ok := <-lda.outboundListener.submitChan:
			if !ok {
				//TODO
				return
			}

			from := outcoming.EmailMessage.Email.SmtpMailFrom[0] //TODO: manage multiple senders
			to := outcoming.EmailMessage.Email.SmtpRcpTo
			var raw bytes.Buffer
			raw.WriteString((&outcoming.EmailMessage.Email.Raw).String())

			// send via local or remote MTA, accordingly
			if outcoming.RemoteCredentials != nil {
				server := strings.Split(outcoming.RemoteCredentials.Host, ":")
				port, _ := strconv.Atoi(server[1])
				var dialErr error
				smtp_remote_sender, dialErr = gomail.NewDialer(
					server[0],
					port,
					outcoming.RemoteCredentials.User,
					outcoming.RemoteCredentials.Password,
				).Dial()
				if dialErr != nil {
					err = fmt.Errorf("outbound: unable to connect to remote MTA with error : %s", dialErr)
				} else {
					err = smtp_remote_sender.Send(from, to, &raw)
				}
			} else {
				if !open {
					var dialErr error
					if smtp_sender, dialErr = d.Dial(); dialErr != nil {
						err = fmt.Errorf("outbound: unable to connect to MTA with error : %s", dialErr)
					} else {
						open = true
					}
				}
				if err == nil {
					err = smtp_sender.Send(from, to, &raw)
				}
			}

			var ack DeliveryAck
			if err != nil {
				log.WithError(err).Warn("outbound: unable to send to MTA")
				ack.Err = true
				ack.Response = err.Error()
			} else {
				ack.Err = false
				ack.Response = ""
			}
			ack.EmailMessage = outcoming.EmailMessage
			outcoming.Response <- &ack
		// Close the connection to the SMTP server and this worker
		// if no email was sent in the last 30 seconds.
		case <-time.After(30 * time.Second):
			if open {
				if err := smtp_sender.Close(); err != nil {
					log.WithError(err).Warn("outbound: unable to close connection to MTA")
				}
				open = false
			}
			return
		}
	}
}

func (c *smtpSender) Send(from string, to []string, msg io.WriterTo) error {
	if err := c.Mail(from); err != nil {
		if err == io.EOF {
			// This is probably due to a timeout, so reconnect and try again.
			sc, derr := c.d.Dial()
			if derr == nil {
				if s, ok := sc.(*smtpSender); ok {
					*c = *s
					return c.Send(from, to, msg)
				}
			}
		}
		return err
	}

	for _, addr := range to {
		if err := c.Rcpt(addr); err != nil {
			return err
		}
	}

	w, err := c.Data()
	if err != nil {
		return err
	}

	if _, err = msg.WriteTo(w); err != nil {
		w.Close()
		return err
	}

	return w.Close()
}
