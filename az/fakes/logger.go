package fakes

type Logger struct {
	PrintlnCall struct {
		CallCount int
		Receives  struct {
			Message string
		}
	}
}

func (l *Logger) Println(message string) {
	l.PrintlnCall.CallCount++
	l.PrintlnCall.Receives.Message = message
}
