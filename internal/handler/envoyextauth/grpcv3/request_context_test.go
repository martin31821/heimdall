// Copyright 2023 Dimitrij Drus <dadrus@gmx.de>
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

package grpcv3

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	corev3 "github.com/envoyproxy/go-control-plane/envoy/config/core/v3"
	envoy_auth "github.com/envoyproxy/go-control-plane/envoy/service/auth/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"

	"github.com/dadrus/heimdall/internal/heimdall"
	"github.com/dadrus/heimdall/internal/heimdall/mocks"
)

func TestNewRequestContext(t *testing.T) {
	t.Parallel()

	// GIVEN
	httpReq := &envoy_auth.AttributeContext_HttpRequest{
		Method:   http.MethodPatch,
		Scheme:   "https",
		Host:     "foo.bar:8080",
		Path:     "/test/baz",
		Query:    "bar=moo",
		Fragment: "foobar",
		Body:     "content=heimdall",
		RawBody:  []byte("content=heimdall"),
		Headers: map[string]string{
			"x-foo-bar":    "barfoo",
			"cookie":       "bar=foo;foo=baz",
			"content-type": "application/x-www-form-urlencoded",
		},
	}
	checkReq := &envoy_auth.CheckRequest{
		Attributes: &envoy_auth.AttributeContext{
			Request: &envoy_auth.AttributeContext_Request{
				Http: httpReq,
			},
		},
	}
	md := metadata.New(nil)
	md.Set("x-forwarded-for", "127.0.0.1", "192.168.1.1")

	ctx := NewRequestContext(
		metadata.NewIncomingContext(
			context.Background(),
			md,
		),
		checkReq,
		mocks.NewJWTSignerMock(t),
	)

	// THEN
	require.Equal(t, httpReq.GetMethod(), ctx.Request().Method)
	require.Equal(t, httpReq.GetScheme(), ctx.Request().URL.Scheme)
	require.Equal(t, httpReq.GetHost(), ctx.Request().URL.Host)
	require.Equal(t, httpReq.GetPath(), ctx.Request().URL.Path)
	require.Equal(t, httpReq.GetFragment(), ctx.Request().URL.Fragment)
	require.Equal(t, httpReq.GetQuery(), ctx.Request().URL.RawQuery)
	require.Equal(t, "moo", ctx.Request().URL.Query().Get("bar"))
	require.Equal(t, httpReq.GetRawBody(), ctx.Request().Body())
	require.Len(t, ctx.Request().Headers(), 3)
	require.Equal(t, "barfoo", ctx.Request().Header("X-Foo-Bar"))
	require.Equal(t, "foo", ctx.Request().Cookie("bar"))
	require.Equal(t, "baz", ctx.Request().Cookie("foo"))
	require.Empty(t, ctx.Request().Cookie("baz"))
	require.NotNil(t, ctx.AppContext())
	require.NotNil(t, ctx.Signer())
	assert.Equal(t, []string{"127.0.0.1", "192.168.1.1"}, ctx.Request().ClientIP)
}

func TestFinalizeRequestContext(t *testing.T) {
	t.Parallel()

	findHeader := func(headers []*corev3.HeaderValueOption, name string) *corev3.HeaderValue {
		for _, header := range headers {
			if header.GetHeader().GetKey() == name {
				return header.GetHeader()
			}
		}

		return nil
	}

	for _, tc := range []struct {
		uc            string
		updateContext func(t *testing.T, ctx heimdall.Context)
		assert        func(t *testing.T, err error, response *envoy_auth.CheckResponse)
	}{
		{
			uc: "successful with some header",
			updateContext: func(t *testing.T, ctx heimdall.Context) {
				t.Helper()

				ctx.AddHeaderForUpstream("x-for-upstream-1", "some-value-1")
				ctx.AddHeaderForUpstream("x-for-upstream-2", "some-value-2")
				ctx.AddHeaderForUpstream("x-for-upstream-1", "some-value-3")
			},
			assert: func(t *testing.T, err error, response *envoy_auth.CheckResponse) {
				t.Helper()

				require.NoError(t, err)
				require.NotNil(t, response)

				assert.Equal(t, int32(codes.OK), response.GetStatus().GetCode())

				okResponse := response.GetOkResponse()
				require.NotNil(t, okResponse)

				require.Len(t, okResponse.GetHeaders(), 2)

				header := findHeader(okResponse.GetHeaders(), "X-For-Upstream-1")
				require.NotNil(t, header)
				assert.Equal(t, "some-value-1,some-value-3", header.GetValue())
				header = findHeader(okResponse.GetHeaders(), "X-For-Upstream-2")
				require.NotNil(t, header)
				assert.Equal(t, "some-value-2", header.GetValue())
			},
		},
		{
			uc: "successful with some cookies",
			updateContext: func(t *testing.T, ctx heimdall.Context) {
				t.Helper()

				ctx.AddCookieForUpstream("some-cookie", "value-1")
				ctx.AddCookieForUpstream("some-other-cookie", "value-2")
			},
			assert: func(t *testing.T, err error, response *envoy_auth.CheckResponse) {
				t.Helper()

				require.NoError(t, err)
				require.NotNil(t, response)

				assert.Equal(t, int32(codes.OK), response.GetStatus().GetCode())

				okResponse := response.GetOkResponse()
				require.NotNil(t, okResponse)

				require.Len(t, okResponse.GetHeaders(), 1)
				assert.Equal(t, "Cookie", okResponse.GetHeaders()[0].GetHeader().GetKey())
				values := strings.Split(okResponse.GetHeaders()[0].GetHeader().GetValue(), ";")
				assert.Len(t, values, 2)
				assert.Contains(t, okResponse.GetHeaders()[0].GetHeader().GetValue(), "some-cookie=value-1")
				assert.Contains(t, okResponse.GetHeaders()[0].GetHeader().GetValue(), "some-other-cookie=value-2")
			},
		},
		{
			uc: "successful with header and cookie",
			updateContext: func(t *testing.T, ctx heimdall.Context) {
				t.Helper()

				ctx.AddHeaderForUpstream("x-for-upstream", "some-value")
				ctx.AddCookieForUpstream("some-cookie", "value-1")
			},
			assert: func(t *testing.T, err error, response *envoy_auth.CheckResponse) {
				t.Helper()

				require.NoError(t, err)
				require.NotNil(t, response)

				assert.Equal(t, int32(codes.OK), response.GetStatus().GetCode())

				okResponse := response.GetOkResponse()
				require.NotNil(t, okResponse)

				require.Len(t, okResponse.GetHeaders(), 2)
				header := findHeader(okResponse.GetHeaders(), "X-For-Upstream")
				require.NotNil(t, header)
				assert.Equal(t, "some-value", header.GetValue())
				header = findHeader(okResponse.GetHeaders(), "Cookie")
				require.NotNil(t, header)
				assert.Equal(t, "some-cookie=value-1", header.GetValue())
			},
		},
		{
			uc: "erroneous with header and cookie",
			updateContext: func(t *testing.T, ctx heimdall.Context) {
				t.Helper()

				ctx.SetPipelineError(errors.New("test error"))
				ctx.AddHeaderForUpstream("x-for-upstream", "some-value")
				ctx.AddCookieForUpstream("some-cookie", "value-1")
				ctx.AddCookieForUpstream("some-other-cookie", "value-2")
			},
			assert: func(t *testing.T, err error, response *envoy_auth.CheckResponse) {
				t.Helper()

				require.Error(t, err)
				assert.Equal(t, "test error", err.Error())
				require.Nil(t, response)
			},
		},
	} {
		t.Run(tc.uc, func(t *testing.T) {
			// GIVEN
			httpReq := &envoy_auth.AttributeContext_HttpRequest{
				Method:   http.MethodPatch,
				Scheme:   "https",
				Host:     "foo.bar:8080",
				Path:     "/test",
				Query:    "bar=moo",
				Fragment: "foobar",
				Body:     "content=heimdall",
				RawBody:  []byte("content=heimdall"),
				Headers: map[string]string{
					"x-foo-bar":    "barfoo",
					"cookie":       "bar=foo;foo=baz",
					"content-type": "application/x-www-form-urlencoded",
				},
			}
			checkReq := &envoy_auth.CheckRequest{
				Attributes: &envoy_auth.AttributeContext{
					Request: &envoy_auth.AttributeContext_Request{
						Http: httpReq,
					},
				},
			}
			ctx := NewRequestContext(context.Background(), checkReq, nil)

			tc.updateContext(t, ctx)

			// WHEN
			resp, err := ctx.Finalize()

			// THEN
			tc.assert(t, err, resp)
		})
	}
}
