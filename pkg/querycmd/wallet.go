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

package querycmd

import (
	"context"
	"database/sql"
	"io"

	"github.com/ericlagergren/decimal"

	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/eupholio"
	"github.com/eupholio/eupholio/pkg/repository"
)

type Balance struct {
	Wallet   string
	Currency string
	Quantity *decimal.Big
}

type WalletBalance struct {
	Wallet []Balance
}

type walletBalance struct {
	Quantity *decimal.Big
}

func QueryWalletBalance(ctx context.Context, w io.Writer, tx *sql.Tx, baseCurrency, source string, of OutputFormat) error {
	repo := repository.New(tx, currency.Symbol(baseCurrency))
	events, err := repo.FindEvents(ctx)
	if err != nil {
		return err
	}

	walletBalances := make(map[string]*walletBalance)

	for _, event := range events {
		currency := event.Currency
		wallet, ok := walletBalances[currency]
		if !ok {
			wallet = &walletBalance{
				Quantity: decimal.New(0, 0),
			}
			walletBalances[currency] = wallet
		}
		switch event.Type {
		case eupholio.EventTypeDeposit:
			wallet.Quantity.Add(wallet.Quantity, event.Quantity.Big)
		case eupholio.EventTypeWithdraw:
			wallet.Quantity.Sub(wallet.Quantity, event.Quantity.Big)
		case eupholio.EventTypeBuy:
			wallet.Quantity.Add(wallet.Quantity, event.Quantity.Big)
		case eupholio.EventTypeSell:
			wallet.Quantity.Sub(wallet.Quantity, event.Quantity.Big)
		case eupholio.EventTypeFee:
			wallet.Quantity.Sub(wallet.Quantity, event.Quantity.Big)
		}
	}
	return nil
}
