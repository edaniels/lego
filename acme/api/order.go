package api

import (
	"context"
	"encoding/base64"
	"errors"

	"github.com/go-acme/lego/v4/acme"
)

type OrderService service

type NewRequest struct {
	Domains []string

	// notAfter (optional, string):
	// The requested value of the notAfter field in the certificate,
	// in the date format defined in [RFC3339].
	NotAfter string
}

// New Creates a new order.
func (o *OrderService) New(ctx context.Context, request NewRequest) (acme.ExtendedOrder, error) {
	var identifiers []acme.Identifier
	for _, domain := range request.Domains {
		identifiers = append(identifiers, acme.Identifier{Type: "dns", Value: domain})
	}

	orderReq := acme.Order{Identifiers: identifiers}
	orderReq.NotAfter = request.NotAfter // defaults to CA's choice

	var order acme.Order
	resp, err := o.core.post(ctx, o.core.GetDirectory().NewOrderURL, orderReq, &order)
	if err != nil {
		return acme.ExtendedOrder{}, err
	}

	return acme.ExtendedOrder{
		Order:    order,
		Location: resp.Header.Get("Location"),
	}, nil
}

// Get Gets an order.
func (o *OrderService) Get(ctx context.Context, orderURL string) (acme.ExtendedOrder, error) {
	if orderURL == "" {
		return acme.ExtendedOrder{}, errors.New("order[get]: empty URL")
	}

	var order acme.Order
	_, err := o.core.postAsGet(ctx, orderURL, &order)
	if err != nil {
		return acme.ExtendedOrder{}, err
	}

	return acme.ExtendedOrder{Order: order}, nil
}

// UpdateForCSR Updates an order for a CSR.
func (o *OrderService) UpdateForCSR(ctx context.Context, orderURL string, csr []byte) (acme.ExtendedOrder, error) {
	csrMsg := acme.CSRMessage{
		Csr: base64.RawURLEncoding.EncodeToString(csr),
	}

	var order acme.Order
	_, err := o.core.post(ctx, orderURL, csrMsg, &order)
	if err != nil {
		return acme.ExtendedOrder{}, err
	}

	if order.Status == acme.StatusInvalid {
		return acme.ExtendedOrder{}, order.Error
	}

	return acme.ExtendedOrder{Order: order}, nil
}
