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

package bitflyer

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/eupholio"
)

type Translator struct {
}

func NewTranslator() *Translator {
	return &Translator{}
}

// Translate stores extracted transaction data to transaction table
func (t *Translator) Translate(ctx context.Context, repo eupholio.Repository, start, end time.Time) error {
	tf := "2006/01/02 15:04:05"
	s := start.Format(tf)
	e := end.Format(tf)

	bitflyerRepository := NewRepository(repo)

	n, err := repo.DeleteTransaction(ctx, WalletCode, start, end)
	if err != nil {
		return err
	}
	if n > 0 {
		log.Println("deleted", n, "entries from events")
	}

	trs, err := bitflyerRepository.FindTransactions(ctx, start, end)
	if err != nil {
		return err
	}
	if len(trs) == 0 {
		log.Println("no transction found:", "start:", s, "end:", e)
	}

	for _, tr := range trs {
		transaction, err := repo.CreateTransaction(ctx, tr.TRDate, WalletCode, tr.ID)
		if err != nil {
			return err
		}
		events, desc, err := t.translateTransaction(transaction, tr)
		if err != nil {
			return err
		}
		transaction.Description = desc
		_, err = transaction.Update(ctx, repo, boil.Infer())
		if err != nil {
			return err
		}

		err = repo.CreateEvents(ctx, events)
		if err != nil {
			return err
		}
	}
	return nil
}

func (t *Translator) translateTransaction(transaction *models.Transaction, tr *models.BFTransaction) (models.EventSlice, string, error) {
	newEvent := eupholio.NewEventFunc(tr.TRDate, transaction.ID)

	var events []*models.Event
	desc := ""
	zero := new(decimal.Big)
	tradingCurrency := tr.Currency1
	paymentCurrency := tr.Currency2.String
	tradingQuantity := tr.Currency1Quantity.Big // trading currency
	feeQuantity := tr.Fee.Big                   // fee in trading currency (negative)
	paymentQuantity := tr.Currency2Quantity.Big // payment currency
	tradingJpyPrice := tr.Currency1JpyRate.Big  // jpy / trading

	jpy := currency.JPY.String()

	switch tr.TRType {
	case TrTypeBuy:
		trading := add(tradingQuantity, feeQuantity)  // position[trading] += trading quantity - fee quantity
		payment := neg(paymentQuantity)               // position[payment] -= payment quantity
		fee := neg(feeQuantity)                       // fee
		cost := mul(tradingJpyPrice, tradingQuantity) // JPY amount equivalent to lost payment quantity
		buy := newEvent(eupholio.EventTypeBuy, tradingCurrency, trading, paymentCurrency, cost)
		sell := newEvent(eupholio.EventTypeSell, paymentCurrency, payment, paymentCurrency, cost)
		commission := newEvent(eupholio.EventTypeCommission, tradingCurrency, feeQuantity, jpy, mul(tradingJpyPrice, fee))
		events = append(events, sell, buy, commission)
		desc = fmt.Sprintf("buy %s/%s", tradingCurrency, paymentCurrency)
	case TrTypeSell:
		trading := neg(add(tradingQuantity, feeQuantity)) // position[trading] -= quote quantity + fee quantity
		payment := paymentQuantity                        // position[payment] += payment quantity
		fee := neg(feeQuantity)                           // fee
		cost := mul(tradingJpyPrice, trading)
		sell := newEvent(eupholio.EventTypeSell, tradingCurrency, trading, jpy, cost)
		buy := newEvent(eupholio.EventTypeBuy, paymentCurrency, payment, jpy, cost)
		commission := newEvent(eupholio.EventTypeCommission, tradingCurrency, fee, jpy, mul(tradingJpyPrice, fee))
		events = append(events, sell, buy, commission)
		desc = fmt.Sprintf("sell %s/%s", tradingCurrency, paymentCurrency)
	case TrTypeReceive:
		trading := tradingQuantity
		f := neg(feeQuantity)
		buy := newEvent(eupholio.EventTypeBuy, tradingCurrency, trading, jpy, zero)
		fee := newEvent(eupholio.EventTypeFee, tradingCurrency, f, jpy, mul(tradingJpyPrice, f))
		events = append(events, buy, fee)
		desc = fmt.Sprintf("receive %s", tr.Currency1)
	case TrTypeTransfer:
		f := neg(feeQuantity)
		withdraw := newEvent(eupholio.EventTypeWithdraw, tradingCurrency, neg(tradingQuantity), jpy, zero)
		fee := newEvent(eupholio.EventTypeFee, tradingCurrency, f, jpy, mul(tradingJpyPrice, f))
		events = append(events, fee, withdraw)
		desc = fmt.Sprintf("transfer %s", tr.Currency1)
	case TrTypeDeposit:
		deposit := newEvent(eupholio.EventTypeDeposit, tr.Currency1, tradingQuantity, jpy, zero)
		events = append(events, deposit)
		desc = fmt.Sprintf("deposit %s", tr.Currency1)
	case TrTypeFee:
		f := neg(tradingQuantity)
		fee := newEvent(eupholio.EventTypeFee, tradingCurrency, f, jpy, mul(tradingJpyPrice, f))
		events = append(events, fee)
		desc = fmt.Sprintf("fee %s", tr.Currency1)
	default:
		log.Println("skip bf type: ", tr.TRType)
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

func neg(x *decimal.Big) *decimal.Big {
	return new(decimal.Big).Neg(x)
}
