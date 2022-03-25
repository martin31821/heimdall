package authenticators

import (
	"github.com/dadrus/heimdall/internal/pipeline/interfaces"
)

type AuthDataGetter interface {
	GetAuthData(s interfaces.AuthDataSource) (string, error)
}
