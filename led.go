package glinet

import "context"

// LEDService calls the led module.
type LEDService service

// LEDConfig is the LED configuration. When Enabled is false, the LEDs are off.
// The LED schedule of TimerService does not change Enabled.
type LEDConfig struct {
	Enabled bool `json:"led_enable"`
}

// GetConfig returns the LED configuration.
func (s *LEDService) GetConfig(ctx context.Context) (*LEDConfig, error) {
	var res LEDConfig
	if err := s.client.Call(ctx, "led", "get_config", nil, &res); err != nil {
		return nil, err
	}

	return &res, nil
}

// SetConfig sets the LED configuration.
func (s *LEDService) SetConfig(ctx context.Context, config LEDConfig) error {
	return s.client.Call(ctx, "led", "set_config", config, nil)
}
