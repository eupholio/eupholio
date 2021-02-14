/*
 * Eupholio - A portfolio tracker tool for cryptocurrency
 * Copyright (C) 2021 Kiyoshi Nakao
 *
 * This file is part of Eupholio.
 *
 * Eupholio is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as
 * published by the Free Software Foundation, either version 3 of the
 * License, or (at your option) any later version.
 *
 * Eupholio is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with Eupholio.  If not, see <http://www.gnu.org/licenses/>.
 */

package poloniex

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/eupholio"
	"github.com/eupholio/eupholio/pkg/poloniex/repository"
	"github.com/eupholio/eupholio/pkg/poloniex/trade"
)

// Translator is a translator for BitTrex
type Translator struct {
	baseCurrency currency.Symbol
}

// NewTranslator create a translator for BitTrex
func NewTranslator(baseCurrency currency.Symbol) *Translator {
	return &Translator{
		baseCurrency: baseCurrency,
	}
}

// Translate stores extracted transaction data to transaction table
func (t *Translator) Translate(ctx context.Context, repo eupholio.Repository, start, end time.Time) error {
	fiat := t.baseCurrency.String()

	poloniexRepository := repository.NewRepository(repo)

	n, err := repo.DeleteTransaction(ctx, WalletCode, start, end)
	if err != nil {
		return err
	}
	if n > 0 {
		log.Println("deleted", n, "entries from events")
	}

	trs, err := poloniexRepository.FindTrades(ctx, start, end)
	if err != nil {
		return err
	}
	if len(trs) == 0 {
		log.Println("no transction found")
	}

	var events []*models.Event

	for _, tr := range trs {
		transaction, err := repo.CreateTransaction(ctx, tr.Date, WalletCode, tr.ID)
		if err != nil {
			return err
		}

		es, desc, err := t.translateTransaction(ctx, repo, transaction, tr)
		if err != nil {
			return err
		}
		transaction.Description = desc
		if _, err := transaction.Update(ctx, repo, boil.Infer()); err != nil {
			return err
		}

		events = append(events, es...)
	}

	deposits, err := poloniexRepository.FindDeposits(ctx, start, end)
	if err != nil {
		return err
	}
	if len(deposits) == 0 {
		log.Println("no deposit found")
	}

	for _, d := range deposits {
		transaction, err := repo.CreateTransaction(ctx, d.Date, WalletCode+"_D", d.ID)
		if err != nil {
			return err
		}
		newEvent := eupholio.NewEventFunc(d.Date, transaction.ID)
		zero := decimal.New(0, 0)
		events = append(events, newEvent(eupholio.EventTypeDeposit, d.Currency, d.Amount.Big, fiat, zero))
	}

	whs, err := poloniexRepository.FindWithdrawals(ctx, start, end)
	if err != nil {
		return err
	}
	if len(whs) == 0 {
		log.Println("no withdraw found")
	}

	for _, w := range whs {
		transaction, err := repo.CreateTransaction(ctx, w.Date, WalletCode+"_W", w.ID)
		if err != nil {
			return err
		}
		newEvent := eupholio.NewEventFunc(w.Date, transaction.ID)
		zero := decimal.New(0, 0)
		events = append(events, newEvent(eupholio.EventTypeWithdraw, w.Currency, w.Amount.Big, fiat, zero))
		events = append(events, newEvent(eupholio.EventTypeFee, w.Currency, w.FeeDeducted.Big, fiat, zero))
	}

	dists, err := poloniexRepository.FindDistributions(ctx, start, end)
	if err != nil {
		return err
	}
	if len(dists) == 0 {
		log.Println("no deposit found")
	}
	for _, d := range dists {
		transaction, err := repo.CreateTransaction(ctx, d.Date, WalletCode+"_A", d.ID)
		if err != nil {
			return err
		}
		newEvent := eupholio.NewEventFunc(d.Date, transaction.ID)
		zero := decimal.New(0, 0)
		events = append(events, newEvent(eupholio.EventTypeBuy, d.Currency, d.Amount.Big, fiat, zero))
	}

	err = repo.CreateEvents(ctx, events)
	if err != nil {
		return err
	}

	return nil
}

// | column name          | value                        |
// | -------------------- | ---------------------------- |
// | Date                 | Time(2006-01-02 15:04:05)    |
// | Market               | Pair (BASE)/(QUOTE)          |
// | Category             | "Exchange"                   |
// | Type                 | "Sell" / "Buy"               |
// | Price                | BASE price in QUOTE          |
// | Amount               | BASE quantity                |
// | Total                | QUOTE quantity               |
// | Fee                  | percentage of fee            |
// | Order Number         | order number                 |
// | Base Total Less Fee  | increase of QUOTE ?          |
// | Quote Total Less Fee | increase of BASE ?           |
// | Fee Currency         | fee currency                 |
// | Fee Total            | fee in QUOTE                 |

// NOTE:
//   The column name of trading file seems confusing. I'm not sure this is the correct understanding.
//    - QuoteTotalLessFee field means differential of base currency (trading currency)
//    - BaseTotalLessFee field means differential of quote currency (payment currency)
func (t *Translator) translateTransaction(ctx context.Context, repository eupholio.Repository, transaction *models.Transaction, tr *models.PoloniexTrade) (models.EventSlice, string, error) {
	newEvent := eupholio.NewEventFunc(tr.Date, transaction.ID)

	desc := ""
	cs := strings.Split(tr.Market, "/")
	if len(cs) < 2 {
		return nil, "", fmt.Errorf("invalid market field %s", tr.Market)
	}
	tradingCurrency, paymentCurrency := cs[0], cs[1]
	tradingQuantity := tr.Amount.Big
	paymentQuantity := tr.Total.Big
	commissionCurrency := tr.FeeCurrency

	commissionQuantity := new(decimal.Big)
	paymentCommissionQuantity := new(decimal.Big)

	switch commissionCurrency {
	case tradingCurrency:
		commissionQuantity.Sub(tradingQuantity, abs(tr.QuoteTotalLessFee.Big)) // incorrect use of "quote"?
		paymentCommissionQuantity.Mul(tr.Price.Big, commissionQuantity)
	case paymentCurrency:
		commissionQuantity.Sub(paymentQuantity, abs(tr.BaseTotalLessFee.Big)) // incorrect use of "base"?
		paymentCommissionQuantity.Set(commissionQuantity)
	default:
		return nil, "", fmt.Errorf("fee currency %s is neither %s nor %s", commissionCurrency, tradingCurrency, paymentCurrency)
	}

	var events []*models.Event
	switch tr.Type {
	case trade.TypeBuy: // get trading currency by payment currency with commission in payment currency
		trading := abs(tr.QuoteTotalLessFee.Big)                                                                                                 // position[trading] += trading
		payment := paymentQuantity                                                                                                               // position[payment] -= payment
		cost := paymentQuantity                                                                                                                  // cost is the total payment currency lost in this transaction
		buy := newEvent(eupholio.EventTypeBuy, tradingCurrency, trading, paymentCurrency, cost)                                                  // get trading currency
		sell := newEvent(eupholio.EventTypeSell, paymentCurrency, payment, paymentCurrency, cost)                                                // loose payment currency
		commission := newEvent(eupholio.EventTypeCommission, commissionCurrency, commissionQuantity, paymentCurrency, paymentCommissionQuantity) // commission information (no effect)
		events = append(events, sell, buy, commission)
		desc += fmt.Sprintf("buy %s/%s w/ %s", tradingCurrency, paymentCurrency, toString(commissionQuantity, commissionCurrency))
	case trade.TypeSell: // get base currency by quote currency with commission in payment currency
		trading := tradingQuantity                                                                                                               // position[trading] += trading
		payment := abs(tr.BaseTotalLessFee.Big)                                                                                                  // position[payment] -= payment
		cost := paymentQuantity                                                                                                                  // paymentQuantity = tradingQuantity * price
		sell := newEvent(eupholio.EventTypeSell, tradingCurrency, trading, paymentCurrency, cost)                                                // loose trading currency
		buy := newEvent(eupholio.EventTypeBuy, paymentCurrency, payment, paymentCurrency, cost)                                                  // get payment currency
		commission := newEvent(eupholio.EventTypeCommission, commissionCurrency, commissionQuantity, paymentCurrency, paymentCommissionQuantity) // commission information (no effect)
		events = append(events, sell, buy, commission)
		desc += fmt.Sprintf("sell %s/%s w/ %s", tradingCurrency, paymentCurrency, toString(commissionQuantity, commissionCurrency))
	default:
		log.Println("skip order type: ", tr.Type)
	}
	return events, desc, nil
}

func mul(x, y *decimal.Big) *decimal.Big {
	return new(decimal.Big).Mul(x, y)
}

func add(x, y *decimal.Big) *decimal.Big {
	return new(decimal.Big).Add(x, y)
}

func sub(x, y *decimal.Big) *decimal.Big {
	return new(decimal.Big).Sub(x, y)
}

func abs(x *decimal.Big) *decimal.Big {
	return new(decimal.Big).Abs(x)
}

func toString(x *decimal.Big, currency string) string {
	switch currency {
	case "JPY":
		return x.RoundToInt().String() + currency
	default:
		return x.Round(8).String() + currency
	}
}
