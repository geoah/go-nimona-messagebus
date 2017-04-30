package events

import "encoding/json"

func jsonCopy(from, to *Event) {
	bs, _ := json.Marshal(from)
	json.Unmarshal(bs, &to)
}
