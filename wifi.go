package glinet

import "context"

// WifiService calls the wifi module.
type WifiService service

// RadioStatus is an item of the result of wifi get_status: the state of a radio, for example
// Name "radio0" and State "ready". The values are runtime values. The router gives no Channel for
// some radios; then Channel is 0. On the test router, radio1 had no channel.
type RadioStatus struct {
	Name    string `json:"name"`
	State   string `json:"state"`
	Channel int    `json:"channel,omitempty"`
}

// GetStatus returns the state of the radios.
func (s *WifiService) GetStatus(ctx context.Context) ([]RadioStatus, error) {
	var res struct {
		Res []RadioStatus `json:"res"`
	}
	if err := s.client.Call(ctx, "wifi", "get_status", nil, &res); err != nil {
		return nil, err
	}

	return res.Res, nil
}
