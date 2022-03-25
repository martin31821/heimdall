package authenticators

import (
	"testing"

	"github.com/dadrus/heimdall/internal/heimdall"
	"github.com/stretchr/testify/assert"
)

func TestAuthenticateWithAnonymousAuthenticatorWithDefaultSubjectId(t *testing.T) {
	// GIVEN
	a := anonymousAuthenticator{}
	sc := heimdall.SubjectContext{}

	// WHEN
	err := a.Authenticate(nil, nil, &sc)

	// THEN
	assert.NoError(t, err)
	assert.Empty(t, sc.Header)
	assert.NotNil(t, sc.Subject)
	assert.Equal(t, "anonymous", sc.Subject.Id)
}

func TestAuthenticateWithAnonymousAuthenticatorWithCustomSubjectId(t *testing.T) {
	// GIVEN
	subjectId := "anon"
	a := anonymousAuthenticator{Subject: subjectId}
	sc := heimdall.SubjectContext{}

	// WHEN
	err := a.Authenticate(nil, nil, &sc)

	// THEN
	assert.NoError(t, err)
	assert.Empty(t, sc.Header)
	assert.NotNil(t, sc.Subject)
	assert.Equal(t, subjectId, sc.Subject.Id)
}
