package message

type Setup struct {
	Type     string `json:"type"`
	Pid      string `json:"pid"`
	Username string `json:"username"`
}

type Inbound struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

type Outbound struct {
	Type     string `json:"type"`
	Pid      string `json:"pid"`
	Username string `json:"username"`
	Content  string `json:"content"`
	DateTime string `json:"datetime"`
}
