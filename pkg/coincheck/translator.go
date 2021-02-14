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

package coincheck

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/eupholio/eupholio/models"
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

	coincheckRepository := NewRepository(repo)

	n, err := repo.DeleteTransaction(ctx, WalletCode, start, end)
	if err != nil {
		return err
	}
	if n > 0 {
		log.Println("deleted", n, "entries from events")
	}

	trs, err := coincheckRepository.FindHistories(ctx, start, end)
	if err != nil {
		return err
	}
	if len(trs) == 0 {
		log.Println("no transction found:", "start:", s, "end:", e)
	}

	for _, tr := range trs {
		transaction, err := repo.CreateTransaction(ctx, tr.Time, WalletCode, tr.ID)
		if err != nil {
			return err
		}
		events, desc, err := t.translateTransaction(transaction, tr)
		if err != nil {
			return err
		}
		if len(desc) > 0 {
			transaction.Description = desc
			_, err = transaction.Update(ctx, repo, boil.Infer())
			if err != nil {
				return err
			}
		}
		if len(events) > 0 {
			err = repo.CreateEvents(ctx, events)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *Translator) translateTransaction(transaction *models.Transaction, tr *models.CoincheckHistory) (models.EventSlice, string, error) {
	newEvent := eupholio.NewEventFunc(tr.Time, transaction.ID)

	var events []*models.Event
	zero := new(decimal.Big)
	desc := ""

	targetCurrency := tr.TradingCurrency
	targetQuantity := tr.Amount.Big
	feeQuantity := tr.Fee.Big // ?
	fiat := FiatCode

	switch tr.Operation {
	case OperationLimitOrder:
	case OperationCompletedTradingContracts:
		rate, tradingCurrency, paymentCurrency, ok := parseCompletedTradingContracts(tr.Comment) // rate = payment / trading
		if !ok {
			return nil, "", fmt.Errorf("failed to parse: %s", tr.Comment)
		}
		switch targetCurrency {
		case tradingCurrency: // buy BTC (ex. BTC-JPY)
			if feeQuantity != nil && feeQuantity.Sign() > 0 {
				return nil, "", fmt.Errorf("fee is not supported yet")
			} else {
				trading := targetQuantity
				payment := mul(rate, targetQuantity)
				cost := mul(rate, targetQuantity)
				buy := newEvent(eupholio.EventTypeBuy, tradingCurrency, trading, paymentCurrency, cost)
				sell := newEvent(eupholio.EventTypeSell, paymentCurrency, payment, paymentCurrency, cost)
				events = append(events, buy, sell)
				desc = fmt.Sprintf("buy %s/%s", tradingCurrency, paymentCurrency)
			}
		case paymentCurrency: // buy JPY (ex. BTC-JPY)
			if feeQuantity != nil && feeQuantity.Sign() > 0 {
				return nil, "", fmt.Errorf("fee is not supported yet")
			} else {
				trading := new(decimal.Big).Quo(targetQuantity, rate)
				payment := targetQuantity // JPY
				cost := payment
				sell := newEvent(eupholio.EventTypeSell, tradingCurrency, trading, paymentCurrency, cost)
				buy := newEvent(eupholio.EventTypeBuy, paymentCurrency, payment, paymentCurrency, cost)
				events = append(events, sell, buy)
				desc = fmt.Sprintf("sell %s/%s", tradingCurrency, paymentCurrency)
			}
		default:
			return nil, "", fmt.Errorf("invalid trading currency %s for %s-%s", targetCurrency, tradingCurrency, paymentCurrency)
		}
	case OperationReceived:
		deposit := newEvent(eupholio.EventTypeDeposit, targetCurrency, targetQuantity, fiat, zero)
		events = append(events, deposit)
		desc = fmt.Sprintf("received %s", toString(targetQuantity, targetCurrency))
	case OperationSent:
		address, ok := parseSent(tr.Comment)
		withdraw := newEvent(eupholio.EventTypeWithdraw, targetCurrency, neg(sub(targetQuantity, feeQuantity)), fiat, zero)
		events = append(events, withdraw)
		if feeQuantity != nil && feeQuantity.Sign() != 0 {
			fee := newEvent(eupholio.EventTypeFee, targetCurrency, neg(feeQuantity), fiat, zero)
			events = append(events, fee)
		}
		if ok {
			desc = fmt.Sprintf("sent %s to %s", targetCurrency, address[0:7])
		} else {
			desc = fmt.Sprintf("sent %s", targetCurrency)
		}
		if feeQuantity != nil && feeQuantity.Sign() != 0 {
			desc = desc + fmt.Sprintf(" with %s", targetCurrency)
		}
	case OperationBankWithdrawal:
		withdraw := newEvent(eupholio.EventTypeWithdraw, targetCurrency, neg(sub(targetQuantity, feeQuantity)), fiat, zero)
		events = append(events, withdraw)
		if feeQuantity != nil && feeQuantity.Sign() != 0 {
			fee := newEvent(eupholio.EventTypeFee, targetCurrency, neg(feeQuantity), fiat, neg(feeQuantity))
			events = append(events, fee)
		}
		desc = fmt.Sprintf("withdraw %s to bank", targetCurrency)
	case OperationCancelLimitOrder:
	default:
		log.Println("skip type: ", tr.Operation)
	}
	return events, desc, nil
}

func mul(x, y *decimal.Big) *decimal.Big {
	return new(decimal.Big).Mul(x, y)
}

func sub(x, y *decimal.Big) *decimal.Big {
	return new(decimal.Big).Sub(x, y)
}

func neg(x *decimal.Big) *decimal.Big {
	return new(decimal.Big).Neg(x)
}

func toString(x *decimal.Big, currency string) string {
	switch currency {
	case "JPY":
		return x.RoundToInt().String() + currency
	default:
		return x.Round(8).String() + currency
	}
}
