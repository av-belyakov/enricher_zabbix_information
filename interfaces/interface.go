package interfaces

//**************** счётчик *****************

type Counter interface {
	SendMessage(string, int)
}

//************** логирование ***************

type Logger interface {
	GetChan() <-chan Messager
	Send(msgType, msgData string)
}

type Messager interface {
	GetType() string
	SetType(v string)
	GetMessage() string
	SetMessage(v string)
}

type WriterLoggingData interface {
	Write(typeLogFile, str string) bool
}
