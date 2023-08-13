package p2p

import (
	"fmt"
	"io"
)

type Handler interface {
	HandleMessage(*Message) error
	Version() string
}

type DefaultHandler struct {
	version string
}

func (h *DefaultHandler) HandleMessage(msg *Message) error {
	b, err := io.ReadAll(msg.Payload)
	if err != nil {
		return err
	}

	fmt.Printf("handling the msg from %s: %s", msg.From, string(b))

	return nil
}

func (h *DefaultHandler) Version() string {
	return h.version
}
