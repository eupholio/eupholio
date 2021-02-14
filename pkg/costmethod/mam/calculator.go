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

package mam

import (
	"fmt"
	"log"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/costmethod"
	"github.com/eupholio/eupholio/pkg/eupholio"
)

type Calculator struct{}

func NewCalculator() *Calculator {
	return &Calculator{}
}

func (cal *Calculator) CalculateBalance(beginingBalances models.BalanceSlice, entries models.EntrySlice, year int, options ...costmethod.Option) (models.BalanceSlice, error) {
	config := &costmethod.Config{}
	for _, o := range options {
		o(config)
	}

	aggregation := make(map[string]*aggregationContext)
	beginings := make(map[string]*models.Balance)

	for _, b := range beginingBalances {
		beginings[b.Currency] = b
		aggregation[b.Currency] = NewAggregationContext(b.Price.Big, b.Quantity.Big)
	}

	for _, entry := range entries {
		ac, ok := aggregation[entry.Currency]
		if !ok {
			ac = NewAggregationContext(decimal.New(0, 0), decimal.New(0, 0))
			aggregation[entry.Currency] = ac
		}
		switch entry.Type {
		case eupholio.EntryTypeOpen:
			ac.ProcessOpen(entry.Quantity.Big, entry.FiatQuantity.Big)
		case eupholio.EntryTypeClose:
			ac.ProcessClose(entry.Quantity.Big, entry.FiatQuantity.Big)
			entry.Price = types.NewNullDecimal(ac.Price())
		default:
			return nil, fmt.Errorf("unknown entry type: %s", entry.Type)
		}
		entry.Price = types.NewNullDecimal(ac.Price())
	}

	var ret models.BalanceSlice
	for currency, ac := range aggregation {
		balance := &models.Balance{
			Year:              year,
			Currency:          currency,
			BeginningQuantity: types.NewDecimal(ac.Beginning()),
			OpenQuantity:      types.NewDecimal(ac.OpenQuantity()),
			CloseQuantity:     types.NewDecimal(ac.CloseQuantity()),
			Price:             types.NewDecimal(ac.Price()),
			Quantity:          types.NewDecimal(ac.Quantity()),
			Profit:            types.NewDecimal(ac.Profit()),
		}
		ret = append(ret, balance)

		if config.Debug {
			log.Println(balanceToString(balance))
		}
	}
	return ret, nil
}

func balanceToString(b *models.Balance) string {
	return fmt.Sprintf("%d %s beginning=%v open=%v close=%v price=%v quantity=%v profit=%v",
		b.Year, b.Currency, b.BeginningQuantity, b.OpenQuantity, b.CloseQuantity, b.Price, b.Quantity, b.Profit)
}
