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

package currency

import (
	"context"
	"database/sql"

	"github.com/eupholio/eupholio/models"
)

type Symbol string

func (t Symbol) String() string {
	return string(t)
}

type SymbolSlice []Symbol

func (ss SymbolSlice) Strings() []string {
	var symbols []string
	for _, c := range ss {
		symbols = append(symbols, c.String())
	}
	return symbols
}

// Fiat
const (
	USD Symbol = "USD"
	EUR Symbol = "EUR"
	JPY Symbol = "JPY"
)

// Crypto
const (
	BTC  Symbol = "BTC"
	ETH  Symbol = "ETH"
	USDT Symbol = "USDT"
	XRP  Symbol = "XRP"
)

var Currencies SymbolSlice

func init() {
	Currencies = append(Currencies, FiatCurrencies...)
	Currencies = append(Currencies, BaseCurrencies...)
}

var FiatCurrencies = SymbolSlice{
	USD, EUR, JPY,
}

var BaseCurrencies = SymbolSlice{
	BTC, ETH, USDT, XRP,
}

var AllCurrencies SymbolSlice

func InitSymbols(ctx context.Context, tx *sql.Tx) error {
	symbols, err := models.Symbols().All(ctx, tx)
	if err != nil {
		return err
	}
	var allCurrencies []Symbol
	for _, s := range symbols {
		allCurrencies = append(allCurrencies, Symbol(s.Symbol))
	}
	AllCurrencies = allCurrencies
	return nil
}

func (s Symbol) IsFiat() bool {
	switch s {
	case USD, EUR, JPY:
		return true
	}
	return false
}
