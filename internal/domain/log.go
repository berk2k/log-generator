package domain

type LogMessage struct {
	Msg   string `json:"msg"`
	Level string `json:"level"`
	Ts    int64  `json:"ts"`
}
