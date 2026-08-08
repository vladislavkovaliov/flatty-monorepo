package applications

import "context"

type Repository interface {
	ListByEnv(ctx context.Context, env string) ([]*Application, error)
	ListAll(ctx context.Context) ([]*Application, error)
	Create(ctx context.Context, input *ApplicationInput) (*Application, error)
	Update(ctx context.Context, id int64, input *ApplicationInput) (*Application, error)
	Delete(ctx context.Context, id int64) (int64, error)
}

type ApplicationInput struct {
	name          string
	env           string
	bundleJS      string
	styleURL      string
	remoteOrigin  string
	proxyBasePath string
	basePath      string
}

func (i *ApplicationInput) Name() string          { return i.name }
func (i *ApplicationInput) Env() string           { return i.env }
func (i *ApplicationInput) BundleJS() string      { return i.bundleJS }
func (i *ApplicationInput) StyleURL() string      { return i.styleURL }
func (i *ApplicationInput) RemoteOrigin() string  { return i.remoteOrigin }
func (i *ApplicationInput) ProxyBasePath() string { return i.proxyBasePath }
func (i *ApplicationInput) BasePath() string      { return i.basePath }

func NewApplicationInput(
	name, env, bundleJS, styleURL, remoteOrigin, proxyBasePath, basePath string,
) *ApplicationInput {
	return &ApplicationInput{
		name:          name,
		env:           env,
		bundleJS:      bundleJS,
		styleURL:      styleURL,
		remoteOrigin:  remoteOrigin,
		proxyBasePath: proxyBasePath,
		basePath:      basePath,
	}
}
