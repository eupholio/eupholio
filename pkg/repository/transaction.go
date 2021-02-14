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

package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/eupholio/eupholio/models"
)

// Transaction

func (r *repository) DeleteTransaction(ctx context.Context, walletCode string, start, end time.Time) (int64, error) {
	s := start.Format(timeFormat)
	e := end.Format(timeFormat)
	n, err := models.Transactions(
		qm.Where("wallet_code = ? AND time >= ? AND time < ?", walletCode, s, e),
	).DeleteAll(ctx, r.ContextExecutor)
	if err != nil {
		return -1, err
	}
	return n, err
}

func (r *repository) CreateTransaction(ctx context.Context, time time.Time, walletCode string, walletTid int) (*models.Transaction, error) {
	transaction := &models.Transaction{
		Time:       time,
		WalletCode: walletCode,
		WalletTid:  walletTid,
	}
	return transaction, transaction.Insert(ctx, r.ContextExecutor, boil.Infer())
}

func (r *repository) FindTransactionsByYear(ctx context.Context, year int, loc *time.Location) (models.TransactionSlice, error) {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, loc).UTC()
	end := time.Date(year+1, time.January, 1, 0, 0, 0, 0, loc).UTC()
	s := start.Format(timeFormat)
	e := end.Format(timeFormat)
	es, err := models.Transactions(
		qm.Where("time >= ? AND time < ?", s, e),
		qm.OrderBy("time ASC, id ASC"),
	).All(ctx, r.ContextExecutor)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return es, err
}
