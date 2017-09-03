package messagebus

type Payload struct {
	Creator string `json:"creator,omitempty"`
	Type    string `json:"type,omitempty"`
	Coded   string `json:"codec,omitempty"`
	Data    []byte `json:"data,omitempty"`
}

type Message struct {
	Payload    Payload `json:"-"`
	PayloadRaw []byte  `json:"payload,omitempty"`
	Signature  []byte  `json:"signature,omitempty"`
}

type Envelope struct {
	Recipipent string  `json:"-"`
	Message    Message `json:"-"`
	MessageRaw []byte  `json:"message,omitempty"`
	Hash       []byte  `json:"hash,omitempty"`
}
