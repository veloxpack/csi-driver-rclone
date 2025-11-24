// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ntlmssp

import (
	"encoding/base64"
	"strings"
)

type authheader []string

func (h authheader) IsBasic() bool {
	for _, s := range h {
		if strings.HasPrefix(s, "Basic ") {
			return true
		}
	}
	return false
}

func (h authheader) Basic() string {
	for _, s := range h {
		if strings.HasPrefix(s, "Basic ") {
			return s
		}
	}
	return ""
}

func (h authheader) IsNegotiate() bool {
	for _, s := range h {
		if strings.HasPrefix(s, "Negotiate") {
			return true
		}
	}
	return false
}

func (h authheader) IsNTLM() bool {
	for _, s := range h {
		if strings.HasPrefix(s, "NTLM") {
			return true
		}
	}
	return false
}

func (h authheader) GetData() ([]byte, error) {
	for _, s := range h {
		if strings.HasPrefix(s, "NTLM") || strings.HasPrefix(s, "Negotiate") || strings.HasPrefix(s, "Basic ") {
			p := strings.Split(s, " ")
			if len(p) < 2 {
				return nil, nil
			}
			return base64.StdEncoding.DecodeString(p[1])
		}
	}
	return nil, nil
}

func (h authheader) GetBasicCreds() (username, password string, err error) {
	d, err := h.GetData()
	if err != nil {
		return "", "", err
	}
	parts := strings.SplitN(string(d), ":", 2)
	return parts[0], parts[1], nil
}
