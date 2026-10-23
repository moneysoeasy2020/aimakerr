package config

import (
	"context"
	"sync"

	"golang.org/x/oauth2"
)

type configTokenSource struct {
	ctx          context.Context
	oauth2Config *oauth2.Config

	readToken  func() (*oauth2.Token, error)
	writeToken func(t *oauth2.Token) error

	// Token is not thread-safe
	mu sync.Mutex
}

func NewConfigTokenSource(ctx context.Context, oauth2Cfg *oauth2.Config, readToken func() (*oauth2.Token, error), writeToken func(t *oauth2.Token) error) (oauth2.TokenSource, error) {
	src := &configTokenSource{
		ctx:          ctx,
		oauth2Config: oauth2Cfg,
		readToken:    readToken,
		writeToken:   writeToken,
	}

	token, err := src.readToken()
	if err != nil {
		return nil, err
	}

	// TODO: double check if ReuseTokenSource is really needed, I think the config already takes care of that.
	return oauth2.ReuseTokenSource(token, src), nil
}

func (c *configTokenSource) Token() (*oauth2.Token, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	token, err := c.readToken()
	if err != nil {
		return nil, err
	}

	refreshedToken, err := c.oauth2Config.TokenSource(c.ctx, token).Token()
	if err != nil {
		return nil, err
	}

	err = c.writeToken(refreshedToken)
	if err != nil {
		return nil, err
	}

	return refreshedToken, nil
}
