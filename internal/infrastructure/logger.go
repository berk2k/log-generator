package infrastructure

import (
	"encoding/json"
	"fmt"
	"log-generator/internal/domain"
)

type ConsoleLogger struct{}

func NewConsoleLogger() *ConsoleLogger {
	return &ConsoleLogger{}
}

func (c *ConsoleLogger) Write(logMsg domain.LogMessage) error {
	b, err := json.Marshal(logMsg)
	if err != nil {
		return err
	}

	fmt.Println(string(b))
	return nil
}
