// Copyright © 2022 Attestant Limited.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package rest

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"

	"github.com/pkg/errors"
)

// Delay holds information about a delay.
type Delay struct {
	IPAddr  *net.IP
	Source  string
	Method  string
	Slot    uint32
	DelayMS uint32
}

// delayJSON is a raw representation of the struct.
type delayJSON struct {
	IPAddr  string `json:"ip_addr,omitempty"`
	Source  string `json:"source"`
	Method  string `json:"method"`
	Slot    string `json:"slot"`
	DelayMS string `json:"delay_ms"`
}

// MarshalJSON implements json.Marshaler.
func (d *Delay) MarshalJSON() ([]byte, error) {
	ipAddr := ""
	if d.IPAddr != nil {
		ipAddr = d.IPAddr.String()
	}

	return json.Marshal(&delayJSON{
		IPAddr:  ipAddr,
		Source:  d.Source,
		Method:  d.Method,
		Slot:    fmt.Sprintf("%d", d.Slot),
		DelayMS: fmt.Sprintf("%d", d.DelayMS),
	})
}

// UnmarshalJSON implements json.Unmarshaler.
func (d *Delay) UnmarshalJSON(input []byte) error {
	var data delayJSON
	err := json.Unmarshal(input, &data)
	if err != nil {
		return err
	}

	if data.IPAddr != "" {
		ipAddr := net.ParseIP(data.IPAddr)
		if ipAddr.To4() != nil {
			ipAddr = ipAddr.To4()
		}
		d.IPAddr = &ipAddr
	}

	if data.Source == "" {
		return errors.New("source missing")
	}
	d.Source = data.Source

	if data.Method == "" {
		return errors.New("method missing")
	}
	d.Method = data.Method

	if data.Slot == "" {
		return errors.New("slot missing")
	}
	slot, err := strconv.ParseUint(data.Slot, 10, 32)
	if err != nil {
		return errors.Wrap(err, "invalid value for slot")
	}
	d.Slot = uint32(slot)

	if data.DelayMS == "" {
		return errors.New("delay_ms missing")
	}
	delayMS, err := strconv.ParseUint(data.DelayMS, 10, 32)
	if err != nil {
		return errors.Wrap(err, "invalid value for delay_ms")
	}
	d.DelayMS = uint32(delayMS)

	return nil
}
