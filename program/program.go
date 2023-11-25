package program

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"flag"
	"fmt"
	"log"

	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"mutualTLS-proxy/pkg/cert"
	pkgContext "mutualTLS-proxy/pkg/context"
	pkgServerLog "mutualTLS-proxy/pkg/server/log"

	gmhttp "github.com/onuryurdupak/gomod/v2/http"
	gmrouting "github.com/onuryurdupak/gomod/v2/http/routing"
	gmlog "github.com/onuryurdupak/gomod/v2/log"
	gmsession "github.com/onuryurdupak/gomod/v2/session"
	"github.com/onuryurdupak/gomod/v2/slice"
)

const (
	VerboseLoggingEnabled = "1"

	EnvServeAddr          = "SERVE_ADDR"
	EnvPathServerKeyFile  = "PATH_SERVER_KEY_FILE"
	EnvPathServerCertFile = "PATH_SERVER_CERT_FILE"
	EnvDirClientCaFiles   = "DIR_CLIENT_CA_FILES"
	EnvRouteBaseAddr      = "ROUTE_BASE_ADDR"
	EnvGatewayTimeoutSecs = "GATEWAY_TIMEOUT_SECS"
	EnvAllowedHttpVerbs   = "ALLOWED_HTTP_VERBS"
	EnvVerboseLogging     = "VERBOSE_LOGGING"
)

var (
	enabledVerboseLogging bool
)

func Main(args []string) {
	if len(args) == 1 {
		switch {
		case args[0] == "version" || args[0] == "--version" || args[0] == "-v":
			fmt.Println(versionInfo())
			os.Exit(ErrSuccess)
			return
		case args[0] == "help" || args[0] == "--help" || args[0] == "-h":
			fmt.Println(helpMessageStyled())
			os.Exit(ErrSuccess)
		case args[0] == "-hr":
			fmt.Println(helpMessageUnstyled())
			os.Exit(ErrSuccess)
		case args[0] == "--raw":
			_ = slice.RemoveString(&args, "--raw")
		default:
			fmt.Println(helpPrompt)
			os.Exit(ErrInput)
		}
	}

	_ = flag.CommandLine.Parse([]string{})

	envVars := []string{
		EnvServeAddr,
		EnvPathServerKeyFile,
		EnvPathServerCertFile,
		EnvDirClientCaFiles,
		EnvRouteBaseAddr,
		EnvGatewayTimeoutSecs,
		EnvAllowedHttpVerbs,
		EnvVerboseLogging,
	}

	err := gmlog.Init("mutualTLS-proxy", "/var/log/mutualTLS-proxy")
	if err != nil {
		log.Fatalf("unable to initialize logger: %s", err.Error())
	}

	sessionID := gmsession.NewID()
	loggerMain := gmlog.NewLogger("Main", sessionID)

	loggerMain.Infof("Starting application.")
	loggerMain.Infof("Reading environment variables.")
	for _, path := range envVars {
		loggerMain.Infof("%s: %s", path, os.Getenv(path))
	}

	enabledVerboseLogging = os.Getenv(EnvVerboseLogging) == VerboseLoggingEnabled

	contextManager := pkgContext.NewContextManager()
	certificateCollector := cert.NewCertificateCollector(contextManager)

	mainCtx := contextManager.InjectSessionID(context.Background(), sessionID)
	certPool := x509.NewCertPool()

	err = certificateCollector.AppendCertsFromDir(mainCtx, certPool, os.Getenv(EnvDirClientCaFiles))
	if err != nil {
		loggerMain.Errorf("Append certificates from directory: %s", err.Error())
		os.Exit(2)
	}

	tlsConfig := &tls.Config{
		ClientCAs:  certPool,
		ClientAuth: tls.RequireAndVerifyClientCert,
	}

	server := &http.Server{
		Addr:      os.Getenv(EnvServeAddr),
		TLSConfig: tlsConfig,
		ErrorLog:  log.New(pkgServerLog.NewServerLogger(), "", 0),
	}

	allowedHttpVerbs := strings.Split(os.Getenv(EnvAllowedHttpVerbs), ";")
	routeRules := make([]*gmrouting.ProxyRouteRule, 0, len(allowedHttpVerbs))

	for _, allowedHttpVerb := range allowedHttpVerbs {
		routeRule := gmrouting.NewProxyRouteRule(allowedHttpVerb, ".*")
		routeRules = append(routeRules, routeRule)
	}

	routeTable, err := gmrouting.NewProxyRouteTable(routeRules)
	if err != nil {
		loggerMain.Errorf("NewProxyRouteTable: %s", err.Error())
		os.Exit(2)
	}

	gatewayTimeoutSecs := os.Getenv(EnvGatewayTimeoutSecs)
	parsedGatewayTimeoutSecs, err := strconv.Atoi(gatewayTimeoutSecs)
	if err != nil {
		loggerMain.Errorf("Unable to parse gateway timeout duration: %s", err.Error())
		os.Exit(2)
	}

	httpClient := &http.Client{
		Timeout: time.Second * time.Duration(parsedGatewayTimeoutSecs),
	}

	responseWriter := gmhttp.NewResponseWriter()

	proxyClient := gmrouting.NewProxyClient(routeTable, os.Getenv(EnvRouteBaseAddr), httpClient, responseWriter, nil,
		// On Error
		func(ctx context.Context, err error) {
			proxySessionID := contextManager.ExtractSessionID(ctx)
			logger := gmlog.NewLogger("ProxyClient", proxySessionID)
			logger.Errorf(err.Error())
		}, nil, nil)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		reqSessionID := gmsession.NewID()
		ctx := contextManager.InjectSessionID(r.Context(), reqSessionID)
		r = r.WithContext(ctx)

		sessionLogger := gmlog.NewLogger("HandleFunc", reqSessionID)

		if enabledVerboseLogging {
			if r.TLS != nil && len(r.TLS.PeerCertificates) > 0 {
				for i, peerCert := range r.TLS.PeerCertificates {
					sessionLogger.Infof("Client Cert Chain [%d]: Subject: %v, Issuer: %v", i, peerCert.Subject, peerCert.Issuer)
				}
			} else {
				sessionLogger.Warnf("No client certificate provided in the request.")
			}
		}

		proxyClient.HandleRequestAndRedirect(w, r)
	})

	loggerMain.Infof("Starting server on %s", server.Addr)

	signalChannel := make(chan os.Signal, 2)
	signal.Notify(signalChannel, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-signalChannel
		switch sig {
		case syscall.SIGINT, syscall.SIGTERM:
			loggerMain.Infof("Received signal: '%v'", sig)

			gmlog.FlushLogger()
			os.Exit(0)
		}
	}()

	err = server.ListenAndServeTLS(os.Getenv(EnvPathServerCertFile), os.Getenv(EnvPathServerKeyFile))
	if err != nil {
		loggerMain.Errorf("ListenAndServeTLS: %s", err.Error())
		os.Exit(2)
	}
}
