// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ntlmssp

import (
	"bytes"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
)

// GetDomain extracts the domain from the username if present.
func GetDomain(username string) (user string, domain string, domainNeeded bool) {
	if strings.Contains(username, "\\") {
		ucomponents := strings.SplitN(username, "\\", 2)
		domain = ucomponents[0]
		user = ucomponents[1]
		domainNeeded = true
	} else if strings.Contains(username, "@") {
		user = username
		domainNeeded = false
	} else {
		user = username
		domainNeeded = true
	}
	return user, domain, domainNeeded
}

// Negotiator is a http.Roundtripper decorator that automatically
// converts basic authentication to NTLM/Negotiate authentication when appropriate.
type Negotiator struct{ http.RoundTripper }

// RoundTrip sends the request to the server, handling any authentication
// re-sends as needed.
func (l Negotiator) RoundTrip(req *http.Request) (res *http.Response, err error) {
	// Use default round tripper if not provided
	rt := l.RoundTripper
	if rt == nil {
		rt = http.DefaultTransport
	}
	// If it is not basic auth, just round trip the request as usual
	reqauth := authheader(req.Header.Values("Authorization"))
	if !reqauth.IsBasic() {
		return rt.RoundTrip(req)
	}
	req = req.Clone(req.Context())
	reqauthBasic := reqauth.Basic()
	// We need to buffer or seek the request body to handle authentication challenges
	// that require resending the body multiple times during the NTLM handshake.
	var body io.ReadSeeker
	var bodyStartPos int64
	if req.Body != nil {
		// Check if body is already seekable to avoid buffering large bodies
		if seeker, ok := req.Body.(io.ReadSeeker); ok {
			// Remember the current position
			bodyStartPos, err = seeker.Seek(0, io.SeekCurrent)
			if err == nil {
				// Seeking succeeded, use the seekable body directly
				body = seeker
				// Close the original body as mandated by http.RoundTripper
				defer req.Body.Close()
			} else {
				// Seeking failed (e.g., pipes), fallback to buffering
				bodyBytes, err := io.ReadAll(req.Body)
				req.Body.Close()
				if err != nil {
					return nil, err
				}
				body = bytes.NewReader(bodyBytes)
				bodyStartPos = 0
			}
		} else {
			// For non-seekable bodies, buffer in memory as required
			bodyBytes, err := io.ReadAll(req.Body)
			req.Body.Close()
			if err != nil {
				return nil, err
			}
			body = bytes.NewReader(bodyBytes)
			bodyStartPos = 0
		}
		req.Body = io.NopCloser(body)
	}
	// first try anonymous, in case the server still finds us
	// authenticated from previous traffic
	req.Header.Del("Authorization")
	res, err = rt.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != http.StatusUnauthorized {
		return res, err
	}
	resauth := authheader(res.Header.Values("Www-Authenticate"))
	if !resauth.IsNegotiate() && !resauth.IsNTLM() {
		// Unauthorized, Negotiate not requested, let's try with basic auth
		req.Header.Set("Authorization", reqauthBasic)
		_, _ = io.Copy(io.Discard, res.Body)
		res.Body.Close()
		if body != nil {
			_, err = body.Seek(bodyStartPos, io.SeekStart)
			if err != nil {
				return nil, err
			}
			req.Body = io.NopCloser(body)
		}

		res, err = rt.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		if res.StatusCode != http.StatusUnauthorized {
			return res, err
		}
		resauth = authheader(res.Header.Values("Www-Authenticate"))
	}

	if resauth.IsNegotiate() || resauth.IsNTLM() {
		// 401 with request:Basic and response:Negotiate
		_, _ = io.Copy(io.Discard, res.Body)
		res.Body.Close()

		// recycle credentials
		u, p, err := reqauth.GetBasicCreds()
		if err != nil {
			return nil, err
		}

		// get domain from username
		domain := ""
		u, domain, domainNeeded := GetDomain(u)

		// send negotiate
		negotiateMessage, err := NewNegotiateMessage(domain, "")
		if err != nil {
			return nil, err
		}
		if resauth.IsNTLM() {
			req.Header.Set("Authorization", "NTLM "+base64.StdEncoding.EncodeToString(negotiateMessage))
		} else {
			req.Header.Set("Authorization", "Negotiate "+base64.StdEncoding.EncodeToString(negotiateMessage))
		}

		if body != nil {
			_, err = body.Seek(bodyStartPos, io.SeekStart)
			if err != nil {
				return nil, err
			}
			req.Body = io.NopCloser(body)
		}

		res, err = rt.RoundTrip(req)
		if err != nil {
			return nil, err
		}

		// receive challenge?
		resauth = authheader(res.Header.Values("Www-Authenticate"))
		challengeMessage, err := resauth.GetData()
		if err != nil {
			return nil, err
		}
		if !(resauth.IsNegotiate() || resauth.IsNTLM()) || len(challengeMessage) == 0 {
			// Negotiation failed, let client deal with response
			return res, nil
		}
		_, _ = io.Copy(io.Discard, res.Body)
		res.Body.Close()

		// send authenticate
		authenticateMessage, err := ProcessChallenge(challengeMessage, u, p, domainNeeded)
		if err != nil {
			return nil, err
		}
		if resauth.IsNTLM() {
			req.Header.Set("Authorization", "NTLM "+base64.StdEncoding.EncodeToString(authenticateMessage))
		} else {
			req.Header.Set("Authorization", "Negotiate "+base64.StdEncoding.EncodeToString(authenticateMessage))
		}

		if body != nil {
			_, err = body.Seek(bodyStartPos, io.SeekStart)
			if err != nil {
				return nil, err
			}
			req.Body = io.NopCloser(body)
		}

		return rt.RoundTrip(req)
	}

	return res, err
}
