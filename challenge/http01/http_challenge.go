package http01

import (
	"context"
	"fmt"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
	"github.com/go-acme/lego/v4/challenge"
)

type ValidateFunc func(ctx context.Context, core *api.Core, domain string, chlng acme.Challenge) error

// ChallengePath returns the URL path for the `http-01` challenge.
func ChallengePath(token string) string {
	return "/.well-known/acme-challenge/" + token
}

type Challenge struct {
	core     *api.Core
	validate ValidateFunc
	provider challenge.Provider
}

func NewChallenge(core *api.Core, validate ValidateFunc, provider challenge.Provider) *Challenge {
	return &Challenge{
		core:     core,
		validate: validate,
		provider: provider,
	}
}

func (c *Challenge) SetProvider(provider challenge.Provider) {
	c.provider = provider
}

func (c *Challenge) Solve(ctx context.Context, authz acme.Authorization) error {
	domain := challenge.GetTargetedDomain(authz)
	c.core.Logger.Infof("[%s] acme: Trying to solve HTTP-01", domain)

	chlng, err := challenge.FindChallenge(challenge.HTTP01, authz)
	if err != nil {
		return err
	}

	// Generate the Key Authorization for the challenge
	keyAuth, err := c.core.GetKeyAuthorization(chlng.Token)
	if err != nil {
		return err
	}

	err = c.provider.Present(authz.Identifier.Value, chlng.Token, keyAuth)
	if err != nil {
		return fmt.Errorf("[%s] acme: error presenting token: %w", domain, err)
	}
	defer func() {
		err := c.provider.CleanUp(authz.Identifier.Value, chlng.Token, keyAuth)
		if err != nil {
			c.core.Logger.Warnf("[%s] acme: cleaning up failed: %v", domain, err)
		}
	}()

	chlng.KeyAuthorization = keyAuth
	return c.validate(ctx, c.core, domain, chlng)
}
