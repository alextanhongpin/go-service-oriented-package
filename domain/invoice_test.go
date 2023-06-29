package domain_test

import (
	"bytes"
	"testing"

	"github.com/alextanhongpin/go-service-oriented-package/domain"
)

func TestInvoice(t *testing.T) {
	inv := domain.NewInvoice()
	inv.Tax = 9
	inv.Add(
		domain.InvoiceLine{
			Description: "red shirt",
			Quantity:    1,
			Price:       1240,
		},
		domain.InvoiceLine{
			Description: "green shirt",
			Quantity:    2,
			Price:       1520,
		},
		domain.InvoiceLine{
			Description: "blue shirt",
			Quantity:    5,
			Price:       905,
		})

	var b bytes.Buffer
	if err := inv.Write(&b); err != nil {
		panic(err)
	}

	t.Log(b.String())
}
