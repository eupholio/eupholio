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

package bittrex

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
func (t *Translator) Translate(ctx context.Context, repository eupholio.Repository, start, end time.Time) error {
	fiat := t.baseCurrency.String()
	s := start.Format(timeFormat)
	e := end.Format(timeFormat)

	bittrexRepository := NewRepository(repository)

	n, err := repository.DeleteTransaction(ctx, WalletCode, start, end)
	if err != nil {
		return err
	}
	if n > 0 {
		log.Println("deleted", n, "entries from events")
	}

	trs, err := bittrexRepository.FindOrderHistories(ctx, start, end)
	if err != nil {
		return err
	}
	if len(trs) == 0 {
		log.Println("no transction found:", "start:", s, "end:", e)
	}

	var events []*models.Event

	for _, tr := range trs {
		transaction, err := repository.CreateTransaction(ctx, tr.Timestamp, WalletCode, tr.ID)
		if err != nil {
			return err
		}

		es, desc, err := t.translateTransaction(ctx, repository, transaction, tr)
		if err != nil {
			return err
		}
		transaction.Description = desc
		if _, err := transaction.Update(ctx, repository, boil.Infer()); err != nil {
			return err
		}

		events = append(events, es...)
	}

	dhs, err := bittrexRepository.FindDepositHistories(ctx, start, end)
	if err != nil {
		return err
	}
	if len(dhs) == 0 {
		log.Println("no deposit found:", "start:", s, "end:", e)
	}

	for _, d := range dhs {
		transaction, err := repository.CreateTransaction(ctx, d.Timestamp, WalletCode+"_D", d.ID)
		if err != nil {
			return err
		}
		newEvent := eupholio.NewEventFunc(d.Timestamp, transaction.ID)
		zero := decimal.New(0, 0)
		events = append(events, newEvent(eupholio.EventTypeDeposit, d.Currency, d.Quantity.Big, fiat, zero))
	}

	whs, err := bittrexRepository.FindWithdrawHistories(ctx, start, end)
	if err != nil {
		return err
	}
	if len(whs) == 0 {
		log.Println("no withdraw found:", "start:", s, "end:", e)
	}

	for _, w := range whs {
		transaction, err := repository.CreateTransaction(ctx, w.Timestamp, WalletCode+"_W", w.ID)
		if err != nil {
			return err
		}
		newEvent := eupholio.NewEventFunc(w.Timestamp, transaction.ID)
		zero := decimal.New(0, 0)
		events = append(events, newEvent(eupholio.EventTypeWithdraw, w.Currency, w.Quantity.Big, fiat, zero))
	}

	err = repository.CreateEvents(ctx, events)
	if err != nil {
		return err
	}

	return nil
}

func (t *Translator) translateTransaction(ctx context.Context, repository eupholio.Repository, transaction *models.Transaction, tr *models.BittrexOrderHistory) (models.EventSlice, string, error) {
	newEvent := eupholio.NewEventFunc(tr.Closed, transaction.ID)

	ss := strings.Split(tr.Exchange, "-")
	if len(ss) < 2 {
		return nil, "", fmt.Errorf("invalid exchange field %s", tr.Exchange)
	}
	paymentCurrency, tradingCurrency := ss[0], ss[1]
	tradingQuantity := tr.Quantity.Big      // trading
	paymentQuantity := tr.Price.Big         // payment (excluding commission)
	commissionQuantity := tr.Commission.Big // commission in payment currency

	desc := ""
	var events []*models.Event
	switch tr.OrderType {
	case OrderTypeLimitBuy:
		trading := tradingQuantity                          // total quantity of trading currency you get
		payment := add(paymentQuantity, commissionQuantity) // total quantity of payment currency you loose
		cost := add(paymentQuantity, commissionQuantity)    // payment currency you loose
		fee := commissionQuantity                           // commission in payment currency
		// add buy/open, sell/close and commission events
		buy := newEvent(eupholio.EventTypeBuy, tradingCurrency, trading, paymentCurrency, cost)          // trading
		sell := newEvent(eupholio.EventTypeSell, paymentCurrency, payment, paymentCurrency, cost)        // payment
		commission := newEvent(eupholio.EventTypeCommission, paymentCurrency, fee, paymentCurrency, fee) // commission
		events = append(events, sell, buy, commission)
		desc = fmt.Sprintf("buy %s/%s", tradingCurrency, paymentCurrency)
	case OrderTypeLimitSell:
		trading := tradingQuantity                          // total quantity of trading currency you loose
		payment := sub(paymentQuantity, commissionQuantity) // total quantity of payment currency you get
		cost := paymentQuantity                             // payment currency equivalent to trading currency you loose
		fee := commissionQuantity                           // fee in payment currency included in trading currency you loose
		// add sell/close, buy/open and commission events
		sell := newEvent(eupholio.EventTypeSell, tradingCurrency, trading, paymentCurrency, cost)        // trading
		buy := newEvent(eupholio.EventTypeBuy, paymentCurrency, payment, paymentCurrency, cost)          // payment
		commission := newEvent(eupholio.EventTypeCommission, paymentCurrency, fee, paymentCurrency, fee) // commission
		events = append(events, sell, buy, commission)
		desc = fmt.Sprintf("sell %s/%s", tradingCurrency, paymentCurrency)
	default:
		log.Println("skip order type: ", tr.OrderType)
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
