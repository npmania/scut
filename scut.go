package scut

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Scut struct {
	MainUrl   string `json:"mainurl"`
	SearchUrl string `json:"searchurl"`
	Key       string `json:"key"`
	UsePOST   bool   `json:"usepost,omitempty"`
}

func Load(f *os.File) (map[string]Scut, error) {
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	var scs map[string]Scut
	err = json.Unmarshal(data, &scs)
	if err != nil {
		return nil, err
	}
	return scs, nil
}
