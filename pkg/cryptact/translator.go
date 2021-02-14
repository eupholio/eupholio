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

package cryptact

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/eupholio"
)

// Translator is a translator for BitTrex
type Translator struct {
	mainCurrency currency.Symbol
}

// NewTranslator create a translator for BitTrex
func NewTranslator(mainCurrency currency.Symbol) *Translator {
	return &Translator{
		mainCurrency: mainCurrency,
	}
}

// Translate stores extracted transaction data to transaction table
func (t *Translator) Translate(ctx context.Context, repo eupholio.Repository, start, end time.Time) error {
	cryptactRepository := NewRepository(repo)

	n, err := repo.DeleteTransaction(ctx, WalletCode, start, end)
	if err != nil {
		return err
	}
	if n > 0 {
		log.Println("deleted", n, "entries from events")
	}

	customs, err := cryptactRepository.FindCustoms(ctx, start, end)
	if err != nil {
		return err
	}
	if len(customs) == 0 {
		log.Println("no transction found")
	}

	var events []*models.Event

	for _, custom := range customs {
		transaction, err := repo.CreateTransaction(ctx, custom.Timestamp, WalletCode, custom.ID)
		if err != nil {
			return err
		}

		es, desc, err := t.TranslateTransaction(ctx, repo, transaction, custom)
		if err != nil {
			return err
		}
		transaction.Description = desc
		if _, err := transaction.Update(ctx, repo, boil.Infer()); err != nil {
			return err
		}

		events = append(events, es...)
	}

	err = repo.CreateEvents(ctx, events)
	if err != nil {
		return err
	}

	return nil
}

func (t *Translator) getValue(ctx context.Context, repository eupholio.Repository, currency string, quantity *decimal.Big, tm time.Time) (*decimal.Big, error) {
	fiat := t.mainCurrency.String()
	valueInFiat := new(decimal.Big)
	switch currency {
	case fiat:
		valueInFiat.Set(quantity)
	default:
		marketPrice, err := repository.FindMarketPriceByCurrencyAndTime(ctx, currency, tm) // base currency price
		if err != nil {
			if err == sql.ErrNoRows {
				return nil, fmt.Errorf("failed to get value because no market price found for %s at %v", currency, tm)
			}
			return nil, err
		}
		valueInFiat.Mul(marketPrice.Price.Big, quantity)
	}
	return valueInFiat, nil
}

// https://support.cryptact.com/hc/en-us/articles/360002571312-Custom-File-for-any-other-trades
//
func (t *Translator) TranslateTransaction(ctx context.Context, repository eupholio.Repository, transaction *models.Transaction, tr *models.CryptactCustom) (models.EventSlice, string, error) {
	newEvent := eupholio.NewEventFunc(tr.Timestamp, transaction.ID)

	desc := ""
	tradingCurrency := tr.Base
	tradingQuantity := tr.Volume.Big
	price := tr.Price.Big // price = payment / trading
	paymentCurrency := tr.Counter
	paymentQuantity := decimal.New(0, 0)
	if price != nil {
		paymentQuantity = mul(price, tradingQuantity)
	}

	// fee
	feeCurrency := tr.FeeCcy
	feeQuantity := tr.Fee.Big

	var events []*models.Event
	switch tr.Action {
	case ActionBuy: // Buy / Hardfork / ICO
		switch feeCurrency {
		case paymentCurrency:
			trading := tradingQuantity                   // position[trading] += trading quantity
			payment := add(paymentQuantity, feeQuantity) // position[payment] -= payment quantity + fee quantity
			cost := add(paymentQuantity, feeQuantity)
			f := feeQuantity
			buy := newEvent(eupholio.EventTypeBuy, tradingCurrency, trading, paymentCurrency, cost)            // trading
			sell := newEvent(eupholio.EventTypeSell, paymentCurrency, payment, paymentCurrency, cost)          // payment
			commission := newEvent(eupholio.EventTypeCommission, feeCurrency, feeQuantity, paymentCurrency, f) // commission
			events = append(events, sell, buy, commission)
		case tradingCurrency:
			trading := sub(tradingQuantity, feeQuantity) // position[trading] += trading quantity - fee quantity
			payment := paymentQuantity                   // position[payment] -= payment quantity
			cost := paymentQuantity
			f := mul(price, feeQuantity)
			buy := newEvent(eupholio.EventTypeBuy, tradingCurrency, trading, paymentCurrency, cost)            // trading
			sell := newEvent(eupholio.EventTypeSell, paymentCurrency, payment, paymentCurrency, cost)          // payment
			commission := newEvent(eupholio.EventTypeCommission, feeCurrency, feeQuantity, paymentCurrency, f) // commission
			events = append(events, sell, buy, commission)
		default:
			return nil, "", fmt.Errorf("fee currency is neither %s nor %s", paymentCurrency, tradingCurrency)
		}
		desc += fmt.Sprintf("buy %s/%s", tradingCurrency, paymentCurrency)
	case ActionSell:
		switch feeCurrency {
		case paymentCurrency:
			trading := tradingQuantity                   // position[trading] -= trading quantity
			payment := sub(paymentQuantity, feeQuantity) // position[payment] += payment quantity - fee quantity
			cost := mul(price, tradingQuantity)
			f := feeQuantity
			sell := newEvent(eupholio.EventTypeSell, tradingCurrency, trading, paymentCurrency, cost)          // trading
			buy := newEvent(eupholio.EventTypeBuy, paymentCurrency, payment, paymentCurrency, cost)            // payment
			commission := newEvent(eupholio.EventTypeCommission, feeCurrency, feeQuantity, paymentCurrency, f) // commission
			events = append(events, sell, buy, commission)
		case tradingCurrency:
			if price == nil {
				return nil, "", fmt.Errorf("price is not specified")
			}
			trading := add(tradingQuantity, feeQuantity) // position[trading] -= trading quantity + fee quantity
			payment := paymentQuantity                   // position[payment] += payment quantity
			cost := mul(price, trading)
			f := mul(price, feeQuantity)
			sell := newEvent(eupholio.EventTypeSell, tradingCurrency, trading, paymentCurrency, cost)          // trading
			buy := newEvent(eupholio.EventTypeBuy, paymentCurrency, payment, paymentCurrency, cost)            // payment
			commission := newEvent(eupholio.EventTypeCommission, feeCurrency, feeQuantity, paymentCurrency, f) // commission
			events = append(events, sell, buy, commission)
		default:
			return nil, "", fmt.Errorf("fee currency is neither %s nor %s", paymentCurrency, tradingCurrency)
		}
		desc += fmt.Sprintf("sell %s/%s w/ %s", tradingCurrency, paymentCurrency, feeCurrency)
	case ActionPay: // Parchese
		if feeCurrency != paymentCurrency {
			return nil, "", fmt.Errorf("fee currency should be %s but %s is specified", paymentCurrency, feeCurrency)
		}
		trading := tradingQuantity // position[trading] -= trading quantity
		if price == nil {
			pay := newEvent(eupholio.EventTypeFee, tradingCurrency, trading, tradingCurrency, tradingQuantity) // pay with "trading currency"
			fee := newEvent(eupholio.EventTypeFee, feeCurrency, feeQuantity, feeCurrency, feeQuantity)         // pay with "trading currency"
			events = append(events, pay, fee)
		} else {
			pay := newEvent(eupholio.EventTypeFee, tradingCurrency, trading, paymentCurrency, paymentQuantity) // pay with specified price (exchange rate)
			fee := newEvent(eupholio.EventTypeFee, feeCurrency, feeQuantity, feeCurrency, feeQuantity)         // pay with "trading currency"
			events = append(events, pay, fee)
		}
	case ActionMining: // Mining
		if feeCurrency != paymentCurrency {
			return nil, "", fmt.Errorf("fee currency should be %s but %s is specified", paymentCurrency, feeCurrency)
		}
		trading := tradingQuantity                                                                // position[trading] += trading quantity
		miningCost := feeQuantity                                                                 // position[fee] -= fee quantity
		buy := newEvent(eupholio.EventTypeBuy, tradingCurrency, trading, feeCurrency, miningCost) // mining
		fee := newEvent(eupholio.EventTypeFee, feeCurrency, miningCost, feeCurrency, miningCost)  // fee
		events = append(events, buy, fee)
		if price == nil {
			sell := newEvent(eupholio.EventTypeSell, tradingCurrency, trading, tradingCurrency, tradingQuantity) // earning
			buy2 := newEvent(eupholio.EventTypeBuy, tradingCurrency, trading, tradingCurrency, tradingQuantity)  // mining
			events = append(events, sell, buy2)
		} else {
			sell := newEvent(eupholio.EventTypeSell, tradingCurrency, trading, paymentCurrency, paymentQuantity) // earning
			buy2 := newEvent(eupholio.EventTypeBuy, tradingCurrency, trading, paymentCurrency, paymentQuantity)  // mining
			events = append(events, sell, buy2)
		}
		desc += fmt.Sprintf("mining %s", tradingCurrency)
	case ActionSendFee:
		if feeCurrency != tradingCurrency {
			desc += "(invalid fee currency)"
		}
		if feeQuantity.Sign() != 0 {
			return nil, "", fmt.Errorf("fee should be 0 but %s is specified", feeQuantity.String())
		}
		if price == nil {
			// position[trading] -= trading quantity
			fee := newEvent(eupholio.EventTypeFee, tradingCurrency, tradingQuantity, tradingCurrency, tradingQuantity) // fee
			events = append(events, fee)
		} else {
			// position[trading] -= trading quantity
			fee := newEvent(eupholio.EventTypeFee, tradingCurrency, tradingQuantity, paymentCurrency, paymentQuantity) // fee
			events = append(events, fee)
		}
	case ActionTip:
	case ActionReduce:
	case ActionBonus:
		buy := newEvent(eupholio.EventTypeBuy, tradingCurrency, tradingQuantity, paymentCurrency, paymentQuantity) // payment is 0 if price is 0
		events = append(events, buy)
	case ActionLending:
		buy := newEvent(eupholio.EventTypeBuy, tradingCurrency, tradingQuantity, paymentCurrency, paymentQuantity) // payment is 0 if price is 0
		events = append(events, buy)
	case ActionStaking:
	default:
		return nil, "", fmt.Errorf("unknown action %s", tr.Action)
	}
	if len(events) == 0 {
		return nil, "", fmt.Errorf("unsuppoted action %s", tr.Action)
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
