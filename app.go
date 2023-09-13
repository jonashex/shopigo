package shopigo

import (
	"fmt"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

var (
	defaultTLDs  = []string{"myshopify.com", "shopify.com", "myshopify.io"}
	subDomainReg = "[a-zA-Z0-9][a-zA-Z0-9-_]*"
)

type App struct {
	*AppConfig
	*Client
	SessionStore
}

func NewAppConfig() *AppConfig {
	return &AppConfig{Credentials: &Credentials{}}
}

type AppConfig struct {
	*Credentials

	HostURL string

	embedded             bool
	authBeginEndpoint    string
	authCallbackPath     string
	authCallbackURL      string
	bypassAuthWithSessID string
	scopes               string
	installHook          func()
	uninstallHookPath    string
	shopRegexp           *regexp.Regexp
}

type Credentials struct {
	ClientID     string
	ClientSecret string
}

func NewApp(c *AppConfig, opts ...Opt) (*App, error) {
	if err := validate(c); err != nil {
		return nil, err
	}
	client, err := NewShopifyClient(&ClientConfig{hostURL: c.HostURL, clientID: c.ClientID})
	if err != nil {
		return nil, err
	}
	app := &App{AppConfig: c, Client: client}
	applyDefaults(app)
	for _, opt := range opts {
		opt(app)
	}
	return app, nil
}

func validate(c *AppConfig) error {
	_, err := url.Parse(c.HostURL)
	return err
}

func applyDefaults(a *App) {
	a.v = VLatest
	a.embedded = true
	a.authBeginEndpoint = "/auth/begin"
	a.authCallbackPath = "/auth/install"
	authCallbackURL, _ := url.JoinPath(a.HostURL, a.authCallbackPath)
	a.authCallbackURL = authCallbackURL
	a.SessionStore = InMemSessionStore
	a.shopRegexp = regexp.MustCompile(fmt.Sprintf("^%s.(%s)/*$", subDomainReg, strings.Join(defaultTLDs, "|")))
}

type Opt = func(a *App)

func WithVersion(v Version) Opt {
	return func(a *App) {
		switch v {
		case V202304:
			a.v = v
		case V202307:
			a.v = v
		default:
			a.v = VLatest
		}
	}
}

func WithRetry(n int) Opt {
	return func(a *App) {
		a.retries = n
	}
}

func WithDefaultAuth(s *Shop) Opt {
	return func(a *App) {
		a.defaultShop = s
	}
}

func WithScopes(s []string) Opt {
	return func(a *App) {
		scopes := make([]string, len(s))
		copy(scopes, s)
		sort.Slice(scopes, func(i, j int) bool {
			return scopes[i] < scopes[j]
		})
		a.scopes = strings.Join(scopes, ",")
	}
}

func WithAuthBeginEndpoint(s string) Opt {
	return func(a *App) {
		a.authBeginEndpoint = s
	}
}

func WithAuthCallbackEndpoint(s string) Opt {
	return func(a *App) {
		a.authCallbackPath = s
		authCallbackURL, err := url.JoinPath(a.HostURL, a.authCallbackPath)
		if err != nil {
			panic(err)
		}
		a.authCallbackURL = authCallbackURL
	}
}

func BypassAuthWithSessionID(s string) Opt {
	return func(a *App) {
		a.bypassAuthWithSessID = s
	}
}

func WithSessionStore(sess SessionStore) Opt {
	return func(a *App) {
		a.SessionStore = sess
	}
}

func WithInstallHook(hook func()) Opt {
	return func(a *App) {
		a.installHook = hook
	}
}

func WithUninstallHook(path string) Opt {
	return func(a *App) {
		a.uninstallHookPath = path
	}
}

func WithIsEmbedded(e bool) Opt {
	return func(a *App) {
		a.embedded = e
	}
}

func WithCustomShopDomains(domains ...string) Opt {
	return func(a *App) {
		a.shopRegexp = regexp.MustCompile(fmt.Sprintf("^%s.(%s)/*$", subDomainReg, strings.Join(append(defaultTLDs, domains...), "|")))
	}
}
