package messages

import (
	"fmt"
	"time"
)

type CLIMessage struct {
	Time       time.Time `json:"time"`
	Command    string    `json:"command"`
	MethodName string    `json:"methodName"`
}

func (m CLIMessage) String() string {
	return fmt.Sprintf(
		"Time: %s; Command: %s; MethodName: %s",
		m.Time.Format(time.DateTime), m.Command, m.MethodName)
}
