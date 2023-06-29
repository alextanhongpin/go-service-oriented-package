package domain

import (
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"text/tabwriter"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

var header = row("DESCRIPTION", "PRICE", "QUANTITY", "AMOUNT")

type Invoice struct {
	InvoiceLine []InvoiceLine
	Tax         int64 // TODO: Change to percent value object
	Printer     func(n int64) string
}

func NewInvoice() *Invoice {
	return &Invoice{
		Printer: defaultPrinter,
	}
}

func (inv *Invoice) Write(out io.Writer) error {
	printer := inv.Printer
	padding := 5

	w := tabwriter.NewWriter(out, 0, 0, padding, ' ', tabwriter.StripEscape|tabwriter.AlignRight|tabwriter.Debug)
	fmt.Fprintln(w, header)

	for _, l := range inv.InvoiceLine {
		fmt.Fprintln(w, row(
			l.Description,
			printer(l.Price),
			strconv.FormatInt(l.Quantity, 10),
			printer(l.Price*l.Quantity),
		))
	}

	if inv.Tax > 0 {
		fmt.Fprintln(w, row("", "", "SUB-TOTAL", printer(inv.SubTotalAmount())))
		fmt.Fprintln(w, row("", "", fmt.Sprintf("%d%% TAX", inv.Tax), printer(inv.TaxAmount())))
	}

	fmt.Fprintln(w, row("", "", "TOTAL", printer(inv.TotalAmount())))
	return w.Flush()
}

func (inv *Invoice) Add(l ...InvoiceLine) {
	inv.InvoiceLine = append(inv.InvoiceLine, l...)
}

func (inv *Invoice) TotalAmount() int64 {
	return inv.SubTotalAmount() + inv.TaxAmount()
}

func (inv *Invoice) SubTotalAmount() int64 {
	var total int64
	for _, line := range inv.InvoiceLine {
		total += line.Quantity + line.Price
	}

	return total
}

func (inv *Invoice) TaxAmount() int64 {
	tax := float64(inv.Tax) / 100.0 * float64(inv.SubTotalAmount())
	tax = math.Round(tax)
	return int64(tax)
}

type InvoiceLine struct {
	Description string
	Price       int64
	Quantity    int64
}

func (l *InvoiceLine) String() string {
	return row(
		l.Description,
		strconv.FormatInt(l.Price, 10),
		strconv.FormatInt(l.Quantity, 10),
		strconv.FormatInt(l.Price*l.Quantity, 10),
	)

}

func defaultPrinter(n int64) string {
	p := message.NewPrinter(language.English)
	return p.Sprintf("%d", n)
}

func row(a, b, c, d string) string {
	s := []string{a, b, c, d}
	return strings.Join(s, "\t") + "\t"
}
