package authenticators

import (
	"errors"
	"testing"

	"github.com/dadrus/heimdall/internal/pipeline/handler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dadrus/heimdall/internal/heimdall"
)

func TestCreateAnonymousAuthenticatorFromValidYaml(t *testing.T) {
	t.Parallel()
	// GIVEN
	conf, err := handler.DecodeTestConfig([]byte("subject: anon"))
	require.NoError(t, err)

	// WHEN
	a, err := newAnonymousAuthenticator(conf)

	// THEN
	assert.NoError(t, err)
	assert.Equal(t, "anon", a.Subject)
}

func TestCreateAnonymousAuthenticatorFromInvalidYaml(t *testing.T) {
	t.Parallel()
	// GIVEN
	conf, err := handler.DecodeTestConfig([]byte("foo: bar"))
	require.NoError(t, err)

	// WHEN
	_, err = newAnonymousAuthenticator(conf)

	// THEN
	assert.Error(t, err)
	assert.True(t, errors.Is(err, heimdall.ErrConfiguration))
}

func TestCreateAnonymousAuthenticatorFromPrototypeGivenEmptyConfig(t *testing.T) {
	t.Parallel()
	// GIVEN
	conf, err := handler.DecodeTestConfig([]byte("subject: anon"))
	require.NoError(t, err)

	prototype, err := newAnonymousAuthenticator(conf)
	assert.NoError(t, err)

	// WHEN
	auth, err := prototype.WithConfig(nil)

	// THEN
	assert.NoError(t, err)

	// prototype and "created" authenticator are same
	assert.Equal(t, prototype, auth)
}

func TestCreateAnonymousAuthenticatorFromPrototypeGivenValidConfig(t *testing.T) {
	t.Parallel()
	// GIVEN
	protoConf, err := handler.DecodeTestConfig([]byte("subject: anon"))
	require.NoError(t, err)

	authConf, err := handler.DecodeTestConfig([]byte("subject: foo"))
	require.NoError(t, err)

	prototype, err := newAnonymousAuthenticator(protoConf)
	assert.NoError(t, err)

	// WHEN
	auth, err := prototype.WithConfig(authConf)

	// THEN
	assert.NoError(t, err)
	// prototype and "created" authenticator are different
	assert.NotEqual(t, prototype, auth)
	aa, ok := auth.(*anonymousAuthenticator)
	require.True(t, ok)
	assert.Equal(t, "foo", aa.Subject)
}

func TestCreateAnonymousAuthenticatorFromPrototypeGivenInvalidConfig(t *testing.T) {
	t.Parallel()
	// GIVEN
	protoConf, err := handler.DecodeTestConfig([]byte("subject: anon"))
	require.NoError(t, err)

	authConf, err := handler.DecodeTestConfig([]byte("foo: bar"))
	require.NoError(t, err)

	prototype, err := newAnonymousAuthenticator(protoConf)
	assert.NoError(t, err)

	// WHEN
	_, err = prototype.WithConfig(authConf)

	// THEN
	assert.Error(t, err)
	assert.True(t, errors.Is(err, heimdall.ErrConfiguration))
}

func TestAuthenticateWithAnonymousAuthenticatorWithCustomSubjectId(t *testing.T) {
	t.Parallel()
	// GIVEN
	subjectID := "anon"
	a := anonymousAuthenticator{Subject: subjectID}

	ctx := &handler.MockContext{}

	// WHEN
	sub, err := a.Authenticate(ctx)

	// THEN
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, subjectID, sub.ID)
	assert.Empty(t, sub.Attributes)
	ctx.AssertExpectations(t)
}

func TestAuthenticateWithAnonymousAuthenticatorWithDefaultSubjectId(t *testing.T) {
	t.Parallel()
	// GIVEN
	auth, err := newAnonymousAuthenticator(nil)
	assert.NoError(t, err)

	ctx := &handler.MockContext{}

	// WHEN
	sub, err := auth.Authenticate(ctx)

	// THEN
	assert.NoError(t, err)
	assert.NotNil(t, sub)
	assert.Equal(t, "anonymous", sub.ID)
	assert.Empty(t, sub.Attributes)
	ctx.AssertExpectations(t)
}
