package extractors

import (
	"strings"

	"github.com/dadrus/heimdall/authenticators"
)

type HeaderValueExtractStrategy struct {
	Name   string
	Prefix string
}

func (es HeaderValueExtractStrategy) GetAuthData(s authenticators.AuthDataSource) (string, error) {
	if val := s.Header(es.Name); len(val) != 0 {
		return strings.TrimSpace(strings.TrimPrefix(val, es.Prefix)), nil
	} else {
		return "", ErrNoAuthDataPresent
	}
}
