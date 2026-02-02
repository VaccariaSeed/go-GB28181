package go_GB28181

import (
	"fmt"

	"github.com/ghettovoice/gosip/sip"
)

var _ sip.Header = (*Subject)(nil)

type Subject struct {
	Str string
}

func (s *Subject) Name() string {
	return "Subject"
}

func (s *Subject) Value() string {
	return s.Str
}

func (s *Subject) Clone() sip.Header {
	return &Subject{s.Str}
}

func (s *Subject) String() string {
	return fmt.Sprintf("Subject: %s", s.Str)
}

func (s *Subject) Equals(other interface{}) bool {
	return true
}
