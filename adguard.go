package glinet

type AdGuardService service

type GetAdGuardConfigResponse struct {
	Enabled bool `json:"enabled"`
}

type SetAdGuardConfigRequest struct {
	Enabled bool `json:"enabled"`
}

func (s *AdGuardService) GetAdGuardConfig() (*GetAdGuardConfigResponse, error) {
	var res GetAdGuardConfigResponse

	err := s.client.CallWithStringSlice("call", []string{s.client.GetSid(), "adguardhome", "get_config"}, &res)
	if err != nil {
		return nil, err
	}

	return &res, nil
}

func (s *AdGuardService) SetAdGuardConfig(Enabled bool) error {
	var req = SetAdGuardConfigRequest{
		Enabled,
	}

	var params = make([]interface{}, 4)
	params[0] = s.client.GetSid()
	params[1] = "adguardhome"
	params[2] = "set_config"
	params[3] = &req

	// don't care about response; it's `[]` if successful
	var res []byte

	err := s.client.CallWithInterfaceSlice("call", params, &res)
	if err != nil {
		return err
	}

	return nil
}
