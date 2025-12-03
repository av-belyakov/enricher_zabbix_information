package helpers

import "github.com/av-belyakov/enricher_zabbix_information/interfaces"

func (l *LoggingForTest) GetChan() <-chan interfaces.Messager {
	return l.chMessage
}

func (l *LoggingForTest) Send(msgType, msgData string) {
	msg := &MessageForTest{}
	msg.SetType(msgType)
	msg.SetMessage(msgData)

	l.chMessage <- msg
}

func (m *MessageForTest) GetType() string {
	return m.msgType
}

func (m *MessageForTest) SetType(v string) {
	m.msgType = v
}

func (m *MessageForTest) GetMessage() string {
	return m.msgData
}

func (m *MessageForTest) SetMessage(v string) {
	m.msgData = v
}
