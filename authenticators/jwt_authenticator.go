package authenticators

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/dadrus/heimdall/config"
	"github.com/dadrus/heimdall/endpoint"
	"github.com/dadrus/heimdall/extractors"
	"github.com/dadrus/heimdall/oauth2"
	"github.com/dadrus/heimdall/pipeline"
	"gopkg.in/square/go-jose.v2"
	"gopkg.in/square/go-jose.v2/jwt"
)

var _ Authenticator = new(jwtAuthenticator)

func newJwtAuthenticator(rawConfig json.RawMessage) (*jwtAuthenticator, error) {
	type _config struct {
		Endpoint       endpoint.Endpoint               `json:"jwks_endpoint"`
		AuthDataSource config.AuthenticationDataSource `json:"jwt_token_from"`
		Assertions     oauth2.Assertions               `json:"jwt_assertions"`
		Session        config.Session                  `json:"session"`
	}

	var c _config
	if err := json.Unmarshal(rawConfig, &c); err != nil {
		return nil, err
	}

	if _, ok := c.Endpoint.Headers["Accept-Type"]; !ok {
		c.Endpoint.Headers["Accept-Type"] = "application/json"
	}
	if len(c.Endpoint.Method) == 0 {
		c.Endpoint.Method = "GET"
	}

	return &jwtAuthenticator{
		e:  c.Endpoint,
		a:  c.Assertions,
		se: c.Session,
		ae: c.AuthDataSource.Strategy(),
	}, nil
}

type jwtAuthenticator struct {
	e  endpoint.Endpoint
	a  oauth2.Assertions
	se config.Session
	ae extractors.AuthDataExtractStrategy
}

func (a *jwtAuthenticator) Authenticate(ctx context.Context, as pipeline.AuthDataSource, sc *pipeline.SubjectContext) error {
	rawBody, err := a.e.SendRequest(ctx, nil)
	if err != nil {
		return err
	}

	var jwks jose.JSONWebKeySet
	if err := json.Unmarshal(rawBody, &jwks); err != nil {
		return err
	}

	jwtRaw, err := a.ae.GetAuthData(as)
	if err != nil {
		return fmt.Errorf("failed to extract jwt: %w", err)
	}

	rawClaims, err := a.verifyTokenAndGetClaims(jwtRaw, jwks)
	if err != nil {
		return err
	}

	if sc.Subject, err = a.se.GetSubject(rawClaims); err != nil {
		return err
	}

	return nil
}

func (a *jwtAuthenticator) verifyTokenAndGetClaims(jwtRaw string, jwks jose.JSONWebKeySet) (json.RawMessage, error) {
	var (
		token *jwt.JSONWebToken
		err   error
	)

	delims := strings.Count(jwtRaw, ".")
	if delims == 2 {
		token, err = jwt.ParseSigned(jwtRaw)
	} else if delims == 3 {
		nestedToken, err := jwt.ParseSignedAndEncrypted(jwtRaw)
		if err != nil {
			return nil, err
		}
		token, err = nestedToken.Decrypt(&jwks)
	} else {
		return nil, errors.New("invalid jwt format")
	}

	if err != nil {
		return nil, err
	}

	var keys []jose.JSONWebKey
	for _, h := range token.Headers {
		keys = jwks.Key(h.KeyID)
		if len(keys) != 0 {
			break
		}
	}
	// even the spec allows for multiple keys for the given id, we do not
	if len(keys) != 1 {
		return nil, errors.New("no (unique) key found for the given key id")
	}

	if !a.a.IsAlgorithmAllowed(keys[0].Algorithm) {
		return nil, fmt.Errorf("%s algorithm is not allowed", keys[0].Algorithm)
	}

	var payload oauth2.JwtPayload
	var tokenClaims map[string]interface{}
	if err = token.Claims(&jwks, &tokenClaims, &payload); err != nil {
		return nil, err
	}

	if err = payload.Verify(a.a); err != nil {
		return nil, err
	}

	rawPayload, err := json.Marshal(tokenClaims)
	if err != nil {
		return nil, err
	}

	return rawPayload, nil
}
