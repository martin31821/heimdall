package admissioncontroller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/inhies/go-bytesize"
	"github.com/justinas/alice"
	"github.com/rs/zerolog"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"

	"github.com/dadrus/heimdall/internal/handler/middleware/http/accesslog"
	"github.com/dadrus/heimdall/internal/handler/middleware/http/dump"
	"github.com/dadrus/heimdall/internal/handler/middleware/http/logger"
	"github.com/dadrus/heimdall/internal/handler/middleware/http/otelmetrics"
	"github.com/dadrus/heimdall/internal/handler/middleware/http/recovery"
	"github.com/dadrus/heimdall/internal/rules/provider/kubernetes/admissioncontroller/admission"
	"github.com/dadrus/heimdall/internal/rules/rule"
	"github.com/dadrus/heimdall/internal/x/httpx"
	"github.com/dadrus/heimdall/internal/x/loggeradapter"
)

type errorHandlerFunc func(http.ResponseWriter, *http.Request, error)

func (f errorHandlerFunc) HandleError(rw http.ResponseWriter, req *http.Request, err error) {
	f(rw, req, err)
}

func newService(
	serviceName string,
	ruleFactory rule.Factory,
	authClass string,
	log zerolog.Logger,
) *http.Server {
	hc := alice.New(
		accesslog.New(log),
		logger.New(log),
		dump.New(),
		recovery.New(errorHandlerFunc(func(rw http.ResponseWriter, _ *http.Request, _ error) {
			rw.WriteHeader(http.StatusInternalServerError)
		})),
		otelhttp.NewMiddleware("",
			otelhttp.WithServerName(serviceName),
			otelhttp.WithSpanNameFormatter(func(_ string, req *http.Request) string {
				return fmt.Sprintf("EntryPoint %s %s%s",
					strings.ToLower(req.URL.Scheme), httpx.LocalAddress(req), req.URL.Path)
			}),
		),
		otelmetrics.New(
			otelmetrics.WithSubsystem("validating admission webhook"),
			otelmetrics.WithServerName(serviceName),
		),
	).Then(newHandler(ruleFactory, authClass))

	return &http.Server{
		Handler:        hc,
		ReadTimeout:    5 * time.Second,      //nolint:gomnd
		WriteTimeout:   10 * time.Second,     //nolint:gomnd
		IdleTimeout:    90 * time.Second,     //nolint:gomnd
		MaxHeaderBytes: int(4 * bytesize.KB), //nolint:gomnd
		ErrorLog:       loggeradapter.NewStdLogger(log),
	}
}

func newHandler(ruleFactory rule.Factory, authClass string) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/validate-ruleset", admission.NewWebhook(&rulesetValidator{f: ruleFactory, ac: authClass}))

	return mux
}
