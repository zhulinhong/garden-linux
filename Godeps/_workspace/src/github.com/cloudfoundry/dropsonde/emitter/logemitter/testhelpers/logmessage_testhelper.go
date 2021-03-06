package testhelpers

import (
	"code.google.com/p/gogoprotobuf/proto"
	"github.com/cloudfoundry/dropsonde/events"
	"time"
)

func NewLogMessage(messageString, appId string) *events.LogMessage {
	messageType := events.LogMessage_OUT
	sourceName := "App"

	return generateLogMessage(messageString, appId, messageType, sourceName, "")
}

func generateLogMessage(messageString, appId string, messageType events.LogMessage_MessageType, sourceName, sourceId string) *events.LogMessage {
	currentTime := time.Now()
	logMessage := &events.LogMessage{
		Message:        []byte(messageString),
		AppId:          proto.String(appId),
		MessageType:    &messageType,
		SourceType:     proto.String(sourceName),
		SourceInstance: proto.String(sourceId),
		Timestamp:      proto.Int64(currentTime.UnixNano()),
	}

	return logMessage
}
