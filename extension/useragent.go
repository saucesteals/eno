package extension

import (
	"encoding/json"

	"github.com/mileusna/useragent"
)

type UserAgent struct {
	useragent.UserAgent
}

func (u *UserAgent) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	u.UserAgent = useragent.Parse(s)
	return nil
}

func (u UserAgent) MarshalJSON() ([]byte, error) {
	return json.Marshal(u.UserAgent.String)
}
