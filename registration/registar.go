package registration

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-acme/lego/v4/acme"
	"github.com/go-acme/lego/v4/acme/api"
)

// Resource represents all important information about a registration
// of which the client needs to keep track itself.
// WARNING: will be remove in the future (acme.ExtendedAccount), https://github.com/go-acme/lego/issues/855.
type Resource struct {
	Body acme.Account `json:"body,omitempty"`
	URI  string       `json:"uri,omitempty"`
}

type RegisterOptions struct {
	TermsOfServiceAgreed bool
}

type RegisterEABOptions struct {
	TermsOfServiceAgreed bool
	Kid                  string
	HmacEncoded          string
}

type Registrar struct {
	core *api.Core
	user User
}

func NewRegistrar(core *api.Core, user User) *Registrar {
	return &Registrar{
		core: core,
		user: user,
	}
}

// Register the current account to the ACME server.
func (r *Registrar) Register(ctx context.Context, options RegisterOptions) (*Resource, error) {
	if r == nil || r.user == nil {
		return nil, errors.New("acme: cannot register a nil client or user")
	}

	accMsg := acme.Account{
		TermsOfServiceAgreed: options.TermsOfServiceAgreed,
		Contact:              []string{},
	}

	if r.user.GetEmail() != "" {
		r.core.Logger.Infof("acme: Registering account for %s", r.user.GetEmail())
		accMsg.Contact = []string{"mailto:" + r.user.GetEmail()}
	}

	account, err := r.core.Accounts.New(ctx, accMsg)
	if err != nil {
		// seems impossible
		var errorDetails acme.ProblemDetails
		if !errors.As(err, &errorDetails) || errorDetails.HTTPStatus != http.StatusConflict {
			return nil, err
		}
	}

	return &Resource{URI: account.Location, Body: account.Account}, nil
}

// RegisterWithExternalAccountBinding Register the current account to the ACME server.
func (r *Registrar) RegisterWithExternalAccountBinding(ctx context.Context, options RegisterEABOptions) (*Resource, error) {
	accMsg := acme.Account{
		TermsOfServiceAgreed: options.TermsOfServiceAgreed,
		Contact:              []string{},
	}

	if r.user.GetEmail() != "" {
		r.core.Logger.Infof("acme: Registering account for %s", r.user.GetEmail())
		accMsg.Contact = []string{"mailto:" + r.user.GetEmail()}
	}

	account, err := r.core.Accounts.NewEAB(ctx, accMsg, options.Kid, options.HmacEncoded)
	if err != nil {
		// seems impossible
		var errorDetails acme.ProblemDetails
		if !errors.As(err, &errorDetails) || errorDetails.HTTPStatus != http.StatusConflict {
			return nil, err
		}
	}

	return &Resource{URI: account.Location, Body: account.Account}, nil
}

// QueryRegistration runs a POST request on the client's registration and returns the result.
//
// This is similar to the Register function,
// but acting on an existing registration link and resource.
func (r *Registrar) QueryRegistration(ctx context.Context) (*Resource, error) {
	if r == nil || r.user == nil || r.user.GetRegistration() == nil {
		return nil, errors.New("acme: cannot query the registration of a nil client or user")
	}

	// Log the URL here instead of the email as the email may not be set
	r.core.Logger.Infof("acme: Querying account for %s", r.user.GetRegistration().URI)

	account, err := r.core.Accounts.Get(ctx, r.user.GetRegistration().URI)
	if err != nil {
		return nil, err
	}

	return &Resource{
		Body: account,
		// Location: header is not returned so this needs to be populated off of existing URI
		URI: r.user.GetRegistration().URI,
	}, nil
}

// UpdateRegistration update the user registration on the ACME server.
func (r *Registrar) UpdateRegistration(ctx context.Context, options RegisterOptions) (*Resource, error) {
	if r == nil || r.user == nil {
		return nil, errors.New("acme: cannot update a nil client or user")
	}

	accMsg := acme.Account{
		TermsOfServiceAgreed: options.TermsOfServiceAgreed,
		Contact:              []string{},
	}

	if r.user.GetEmail() != "" {
		r.core.Logger.Infof("acme: Registering account for %s", r.user.GetEmail())
		accMsg.Contact = []string{"mailto:" + r.user.GetEmail()}
	}

	accountURL := r.user.GetRegistration().URI

	account, err := r.core.Accounts.Update(ctx, accountURL, accMsg)
	if err != nil {
		return nil, err
	}

	return &Resource{URI: accountURL, Body: account}, nil
}

// DeleteRegistration deletes the client's user registration from the ACME server.
func (r *Registrar) DeleteRegistration(ctx context.Context) error {
	if r == nil || r.user == nil {
		return errors.New("acme: cannot unregister a nil client or user")
	}

	r.core.Logger.Infof("acme: Deleting account for %s", r.user.GetEmail())

	return r.core.Accounts.Deactivate(ctx, r.user.GetRegistration().URI)
}

// ResolveAccountByKey will attempt to look up an account using the given account key
// and return its registration resource.
func (r *Registrar) ResolveAccountByKey(ctx context.Context) (*Resource, error) {
	r.core.Logger.Infof("acme: Trying to resolve account by key")

	accMsg := acme.Account{OnlyReturnExisting: true}
	account, err := r.core.Accounts.New(ctx, accMsg)
	if err != nil {
		return nil, err
	}

	return &Resource{URI: account.Location, Body: account.Account}, nil
}
