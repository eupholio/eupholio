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

package costmethod

import (
	"context"
	"fmt"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/eupholio"
)

func CalculateFiatPrice(ctx context.Context, repo eupholio.Repository, year int, loc *time.Location, fiat currency.Symbol) error {
	lastBalances, err := repo.FindBalancesByYear(ctx, year-1)
	if err != nil {
		return err
	}

	transactions, err := eupholio.FindEventsOfTransactions(ctx, repo, year, loc)
	if err != nil {
		return err
	}

	cctx := NewCaluculateContext()
	for _, b := range lastBalances {
		cctx.InitPosition(b.Currency, b.Quantity.Big)
	}

	var newEntries models.EntrySlice
	for _, transaction := range transactions {
		var entrySlice models.EntrySlice
		var open *models.Entry
		var close *models.Entry

		for _, event := range transaction.Events {
			newEntry := func(entryType string, fiatQuantity *decimal.Big) *models.Entry {
				return &models.Entry{
					ID:            event.ID,
					TransactionID: event.TransactionID,
					Time:          event.Time,
					Type:          entryType,
					Currency:      event.Currency,
					Quantity:      types.NewDecimal(event.Quantity.Big),
					Position:      types.NewDecimal(cctx.Position(event.Currency)),
					FiatCurrency:  fiat.String(),
					FiatQuantity:  types.NewDecimal(fiatQuantity),
					Commission:    types.NewNullDecimal(nil),
				}
			}

			fiatQuantity, err := calculateFiatValue(ctx, repo, fiat, event)
			if err != nil {
				return err
			}
			switch event.Type {
			case eupholio.EventTypeBuy:
				cctx.OpenPosition(event.Currency, event.Quantity.Big)
				open = newEntry(eupholio.EntryTypeOpen, fiatQuantity)
				entrySlice = append(entrySlice, open)
			case eupholio.EventTypeSell:
				cctx.ClosePosition(event.Currency, event.Quantity.Big)
				close = newEntry(eupholio.EntryTypeClose, fiatQuantity)
				entrySlice = append(entrySlice, close)
			case eupholio.EventTypeCommission:
				if open != nil && open.Currency == event.Currency {
					open.Commission = types.NewNullDecimal(fiatQuantity)
					open = nil
				} else if close != nil && close.Currency == event.Currency {
					open.Commission = types.NewNullDecimal(fiatQuantity)
					close = nil
				} else {
					return fmt.Errorf("internal error: open or close entry not found")
				}
			case eupholio.EventTypeFee:
				cctx.ClosePosition(event.Currency, event.Quantity.Big)
				fee := newEntry(eupholio.EntryTypeClose, fiatQuantity)
				entrySlice = append(entrySlice, fee)
			}
		}
		newEntries = append(newEntries, entrySlice...)
	}

	if err := repo.CreateEntries(ctx, newEntries); err != nil {
		return err
	}
	return nil
}

func calculateFiatValue(ctx context.Context, repo eupholio.MarketPriceRepository, fiat currency.Symbol, event *models.Event) (*decimal.Big, error) {
	fiatQuantity := new(decimal.Big)
	if event.BaseCurrency != fiat.String() {
		fiat, err := repo.FindMarketPriceByCurrencyAndTime(ctx, event.BaseCurrency, event.Time)
		if err != nil {
			return nil, err
		}
		fiatQuantity.Mul(fiat.Price.Big, event.BaseQuantity.Big)
	} else {
		fiatQuantity.Set(event.BaseQuantity.Big)
	}
	return fiatQuantity, nil
}

// UpdateBalanceByYear calcurates the profit of a year
func UpdateBalanceByYear(ctx context.Context, repo eupholio.Repository, year int, loc *time.Location, fiat currency.Symbol, c Calculator, options ...Option) error {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	end := time.Date(year+1, time.January, 1, 0, 0, 0, 0, loc)

	lastBalances, err := repo.FindBalancesByYear(ctx, year-1)
	if err != nil {
		return err
	}

	entries, err := repo.FindEntriesByStartAndEnd(ctx, start, end)
	if err != nil {
		return err
	}

	bs, err := c.CalculateBalance(lastBalances, entries, year, options...)
	if err != nil {
		return err
	}

	var balances models.BalanceSlice
	for _, b := range bs {
		if b.Currency == fiat.String() {
			continue
		}
		balances = append(balances, b)
	}

	err = repo.CreateBalances(ctx, balances)
	if err != nil {
		return err
	}

	err = repo.UpdateEntries(ctx, entries)
	if err != nil {
		return err
	}
	return nil
}
