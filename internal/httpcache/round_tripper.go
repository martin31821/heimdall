// Copyright 2022 Dimitrij Drus <dadrus@gmx.de>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//
// SPDX-License-Identifier: Apache-2.0

package httpcache

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"

	"github.com/pquerna/cachecontrol"

	"github.com/dadrus/heimdall/internal/cache"
	"github.com/dadrus/heimdall/internal/x/stringx"
)

var (
	ErrInvalidCacheEntry = errors.New("invalid cache entry")
	ErrNoCacheEntry      = errors.New("no cache entry")
)

type RoundTripper struct {
	Transport http.RoundTripper
}

func (rt *RoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := rt.cachedResponse(req)
	if err == nil {
		return resp, nil
	}

	resp, err = rt.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	rt.cacheResponse(req, resp)

	return resp, nil
}

func (rt *RoundTripper) cachedResponse(req *http.Request) (*http.Response, error) {
	cch := cache.Ctx(req.Context())

	cachedValue := cch.Get(cacheKey(req))
	if cachedValue == nil {
		return nil, ErrNoCacheEntry
	}

	respDump, ok := cachedValue.([]byte)
	if !ok {
		return nil, ErrInvalidCacheEntry
	}

	return http.ReadResponse(bufio.NewReader(bytes.NewReader(respDump)), req)
}

func (rt *RoundTripper) cacheResponse(req *http.Request, resp *http.Response) {
	defaultExpirationTime := time.Time{}

	reasons, expires, err := cachecontrol.CachableResponse(req, resp, cachecontrol.Options{PrivateCache: true})
	if err != nil || len(reasons) != 0 || expires == defaultExpirationTime {
		return
	}

	respDump, err := httputil.DumpResponse(resp, true)
	if err != nil {
		return
	}

	cch := cache.Ctx(req.Context())
	cch.Set(cacheKey(req), respDump, time.Until(expires))
}

func cacheKey(req *http.Request) string {
	hash := sha256.New()

	hash.Write(stringx.ToBytes("RFC 7234"))
	hash.Write(stringx.ToBytes(req.URL.String()))
	hash.Write(stringx.ToBytes(req.Method))

	value := req.Header.Get("Authorization")
	if len(value) != 0 {
		hash.Write(stringx.ToBytes(strings.TrimSpace(value)))
	}

	return hex.EncodeToString(hash.Sum(nil))
}
