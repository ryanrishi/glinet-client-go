package glinet

// APIClient manages communication with the GL.iNet API API v3.0.1
// In most cases there should be only one, shared, APIClient.
//type Client struct {
//}

// callAPI do the request.
//func (c *Client) call(method string, req, res interface{}) error {
//	if c.cfg.Debug {
//		dump, err := httputil.DumpRequestOut(request, true)
//		if err != nil {
//			return nil, err
//		}
//		log.Printf("\n%s\n", string(dump))
//	}
//
//	resp, err := c.cfg.HTTPClient.Do(request)
//	if err != nil {
//		return resp, err
//	}
//
//	if c.cfg.Debug {
//		dump, err := httputil.DumpResponse(resp, true)
//		if err != nil {
//			return resp, err
//		}
//		log.Printf("\n%s\n", string(dump))
//	}
//	return resp, err
//}
