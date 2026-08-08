package applications

import "time"

type Application struct {
	id            int64
	name          string
	env           string
	bundleJS      string
	styleURL      string
	remoteOrigin  string
	proxyBasePath string
	basePath      string
	createdAt     time.Time
	updatedAt     time.Time
}

func (a *Application) ID() int64             { return a.id }
func (a *Application) Name() string          { return a.name }
func (a *Application) Env() string           { return a.env }
func (a *Application) BundleJS() string      { return a.bundleJS }
func (a *Application) StyleURL() string      { return a.styleURL }
func (a *Application) RemoteOrigin() string  { return a.remoteOrigin }
func (a *Application) ProxyBasePath() string { return a.proxyBasePath }
func (a *Application) BasePath() string      { return a.basePath }
func (a *Application) CreatedAt() time.Time  { return a.createdAt }
func (a *Application) UpdatedAt() time.Time  { return a.updatedAt }

func NewApplication(
	id int64,
	name, env, bundleJS, styleURL, remoteOrigin, proxyBasePath, basePath string,
	createdAt, updatedAt time.Time,
) *Application {
	return &Application{
		id:            id,
		name:          name,
		env:           env,
		bundleJS:      bundleJS,
		styleURL:      styleURL,
		remoteOrigin:  remoteOrigin,
		proxyBasePath: proxyBasePath,
		basePath:      basePath,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}
