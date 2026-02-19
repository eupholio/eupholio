package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/costmethod/mam"
	"github.com/eupholio/eupholio/pkg/costmethod/wam"
	"github.com/eupholio/eupholio/pkg/eupholio"
)

type Event struct {
	Type        string `json:"type"`
	Asset       string `json:"asset"`
	Qty         string `json:"qty"`
	JpyCost     string `json:"jpy_cost"`
	JpyProceeds string `json:"jpy_proceeds"`
	JpyValue    string `json:"jpy_value"`
}

type Input struct {
	TaxYear int     `json:"tax_year"`
	Events  []Event `json:"events"`
}

type Out struct {
	Method         string            `json:"method"`
	RealizedPnlJpy string            `json:"realized_pnl_jpy"`
	Positions      map[string]string `json:"positions"`
}

func decFromString(s string) *decimal.Big {
	d := new(decimal.Big)
	if _, ok := d.SetString(s); !ok {
		panic("invalid decimal: " + s)
	}
	return d
}

func toDecimal(s string) types.Decimal {
	return types.NewDecimal(decFromString(s))
}

func buildEntries(events []Event) models.EntrySlice {
	ret := models.EntrySlice{}
	for _, e := range events {
		switch e.Type {
		case "Acquire":
			ret = append(ret, &models.Entry{Type: eupholio.EntryTypeOpen, Currency: e.Asset, Quantity: toDecimal(e.Qty), FiatQuantity: toDecimal(e.JpyCost)})
		case "Dispose":
			ret = append(ret, &models.Entry{Type: eupholio.EntryTypeClose, Currency: e.Asset, Quantity: toDecimal(e.Qty), FiatQuantity: toDecimal(e.JpyProceeds)})
		case "Income":
			ret = append(ret, &models.Entry{Type: eupholio.EntryTypeOpen, Currency: e.Asset, Quantity: toDecimal(e.Qty), FiatQuantity: toDecimal(e.JpyValue)})
		case "Transfer":
			// ignored in go parity check (no direct equivalent entry type in cost calculators)
		}
	}
	return ret
}

func calcMAM(year int, entries models.EntrySlice) Out {
	calc := mam.NewCalculator()
	balances, err := calc.CalculateBalance(models.BalanceSlice{}, entries, year)
	if err != nil {
		panic(err)
	}
	profit := decimal.New(0, 0)
	positions := map[string]string{}
	for _, b := range balances {
		profit.Add(profit, b.Profit.Big)
		positions[b.Currency] = b.Quantity.Big.String()
	}
	return Out{Method: "mam", RealizedPnlJpy: profit.String(), Positions: positions}
}

func calcWAM(year int, entries models.EntrySlice) Out {
	calc := wam.NewCalculator()
	balances, err := calc.CalculateBalance(models.BalanceSlice{}, entries, year)
	if err != nil {
		panic(err)
	}
	profit := decimal.New(0, 0)
	positions := map[string]string{}
	for _, b := range balances {
		profit.Add(profit, b.Profit.Big)
		positions[b.Currency] = b.Quantity.Big.String()
	}
	return Out{Method: "wam", RealizedPnlJpy: profit.String(), Positions: positions}
}

func main() {
	var in Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		panic(err)
	}
	entries := buildEntries(in.Events)
	outs := []Out{calcMAM(in.TaxYear, entries), calcWAM(in.TaxYear, entries)}
	if err := json.NewEncoder(os.Stdout).Encode(outs); err != nil {
		panic(err)
	}
	fmt.Println()
}
