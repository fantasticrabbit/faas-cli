// Copyright (c) OpenFaaS Author(s) 2018. All rights reserved.
// Licensed under the MIT license. See LICENSE file in the project root for full license information.

package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"path"

	types "github.com/openfaas/faas-provider/types"
)

//GetFunctionInfo get an OpenFaaS function information
func (c *Client) GetFunctionInfo(ctx context.Context, functionName string, namespace string) (types.FunctionStatus, error) {
	var (
		result types.FunctionStatus
		err    error
	)

	u := c.GatewayURL
	v := u.Query()
	if len(namespace) > 0 {
		v.Set("namespace", namespace)
	}

	// Request CPU/RAM usage if available
	v.Set("usage", "1")

	u.Path = path.Join(functionPath, functionName)
	u.RawQuery = v.Encode()

	req, err := c.newRequestByURL(http.MethodGet, u, nil)
	if err != nil {
		return result, fmt.Errorf("cannot create URL: %s, error: %w", u.String(), err)
	}

	res, err := c.doRequest(ctx, req)
	if err != nil {
		return result, fmt.Errorf("cannot connect to URL: %s, error: %w", c.GatewayURL.String(), err)
	}

	if res.Body != nil {
		defer res.Body.Close()
	}

	switch res.StatusCode {
	case http.StatusOK:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return result, fmt.Errorf("cannot read result from OpenFaaS on URL: %s", c.GatewayURL.String())
		}

		if err := json.Unmarshal(bytesOut, &result); err != nil {
			return result, fmt.Errorf("cannot parse result from OpenFaaS on URL: %s\n%w",
				c.GatewayURL.String(), err)
		}

	case http.StatusUnauthorized:
		return result, fmt.Errorf("unauthorized access, run \"faas-cli login\" to setup authentication for this server")
	case http.StatusNotFound:
		return result, fmt.Errorf("no such function: %s", functionName)
	default:
		bytesOut, err := ioutil.ReadAll(res.Body)
		if err == nil {
			return result, fmt.Errorf("server returned unexpected status code: %d - %s", res.StatusCode, string(bytesOut))
		}
	}
	return result, nil
}
