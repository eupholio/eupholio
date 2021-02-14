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

package wam

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

	positions := costmethod.NewCaluculateContext()
	amounts := costmethod.NewCaluculateContext()

	for _, b := range beginingBalances {
		positions.InitPosition(b.Currency, b.Quantity.Big)
		amount := new(decimal.Big).Mul(b.Price.Big, b.Quantity.Big)
		amounts.InitPosition(b.Currency, amount)
	}

	for _, entry := range entries {
		switch entry.Type {
		case eupholio.EntryTypeOpen:
			pos := positions.Position(entry.Currency)
			positions.OpenPosition(entry.Currency, entry.Quantity.Big)
			if config.Debug {
				log.Print("wam: ", entry.Currency, " ", positions.Position(entry.Currency), " = ", pos.String(), " + ", entry.Quantity.Big.String())
			}
			amounts.OpenPosition(entry.Currency, entry.FiatQuantity.Big)
		case eupholio.EntryTypeClose:
			pos := positions.Position(entry.Currency)
			positions.ClosePosition(entry.Currency, entry.Quantity.Big)
			if config.Debug {
				log.Print("wam: ", entry.Currency, " ", positions.Position(entry.Currency), " = ", pos.String(), " - ", entry.Quantity.Big.String())
			}
			amounts.ClosePosition(entry.Currency, entry.FiatQuantity.Big)
		}
	}

	balances := make(map[string]*models.Balance)
	var ret models.BalanceSlice

	for currency, position := range positions.Balances() {
		amount, _ := amounts.Balance(currency)

		// calculate weighted average price
		// weighted price = (inventory amount + buy amount) / (inventory quantity + buy quantity)
		totalAmount := new(decimal.Big).Add(amount.Init, amount.Open)
		totalQuantity := new(decimal.Big).Add(position.Init, position.Open)
		weightedPrice := decimal.New(0, 0)
		if totalQuantity.Sign() == 1 {
			weightedPrice.Quo(totalAmount, totalQuantity)
		}

		// calculate cost
		// cost amount = sell quantity * weighted price
		costAmount := new(decimal.Big).Mul(position.Close, weightedPrice)

		// calculate profit
		// profit amount = sell amount - cost amount
		profitAmount := new(decimal.Big).Sub(amount.Close, costAmount)

		// calculate resulted quantity
		// quantity = total quantity - sell quantity - fee quantity
		quantity := new(decimal.Big).Sub(totalQuantity, position.Close)

		balance := &models.Balance{
			Year:              year,
			Currency:          currency,
			BeginningQuantity: types.NewDecimal(position.Init),
			OpenQuantity:      types.NewDecimal(position.Open),
			CloseQuantity:     types.NewDecimal(position.Close),
			Price:             types.NewDecimal(weightedPrice),
			Quantity:          types.NewDecimal(quantity),
			Profit:            types.NewDecimal(profitAmount),
		}
		ret = append(ret, balance)
		balances[currency] = balance

		if config.Debug {
			log.Println(balanceToString(position.Init, balance))
		}
	}

	for _, entry := range entries {
		currency := entry.Currency
		b := balances[currency]
		price := new(decimal.Big).Copy(b.Price.Big)
		entry.Price = types.NewNullDecimal(price)
	}
	return ret, nil
}

func newBalance(year int, currency string) *models.Balance {
	zero := func() types.Decimal {
		return types.NewDecimal(decimal.New(0, 0))
	}
	return &models.Balance{
		Year:          year,
		Currency:      currency,
		OpenQuantity:  zero(),
		CloseQuantity: zero(),
		Price:         zero(),
		Quantity:      zero(),
		Profit:        zero(),
	}
}

func balanceToString(begin *decimal.Big, b *models.Balance) string {
	return fmt.Sprintf("%d %s begin=%v open=%v close=%v price=%v quantity=%v profit=%v",
		b.Year, b.Currency, begin, b.OpenQuantity, b.CloseQuantity, b.Price, b.Quantity, b.Profit)
}
