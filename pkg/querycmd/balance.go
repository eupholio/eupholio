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
	"fmt"
	"io"

	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/repository"
)

func QueryBalance(ctx context.Context, writer io.Writer, tx *sql.Tx, year int, fiat currency.Symbol, source string, of OutputFormat) error {
	repo := repository.New(tx, fiat)
	balances, err := repo.FindBalancesByYear(ctx, year)
	if err != nil {
		return err
	}
	switch of {
	case OutputFormatTable:
		NewTableWriter(writer).PrintBalances(balances)
	case OutputFormatCSV:
	default:
		return fmt.Errorf("unknown output format %s", of)
	}
	return nil
}
