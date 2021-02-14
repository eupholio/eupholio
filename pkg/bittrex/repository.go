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

package bittrex

import (
	"context"
	"log"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/eupholio/eupholio/models"
)

const timeFormat = "2006/01/02 15:04:05"

type Repository interface {
	FindOrderHistories(ctx context.Context, start, end time.Time) (models.BittrexOrderHistorySlice, error)
	FindDepositHistories(ctx context.Context, start, end time.Time) (models.BittrexDepositHistorySlice, error)
	FindWithdrawHistories(ctx context.Context, start, end time.Time) (models.BittrexWithdrawHistorySlice, error)
}

func NewRepository(db boil.ContextExecutor) Repository {
	return &repository{
		db: db,
	}
}

type repository struct {
	db boil.ContextExecutor
}

func (r *repository) FindOrderHistories(ctx context.Context, start, end time.Time) (models.BittrexOrderHistorySlice, error) {
	s := start.Format(timeFormat)
	e := end.Format(timeFormat)
	trs, err := models.BittrexOrderHistories(
		qm.Where("timestamp >= ? AND timestamp < ?", s, e),
		qm.OrderBy("timestamp ASC"),
	).All(ctx, r.db)
	if err != nil {
		log.Println("failed to find transactions:", "start:", s, "end:", e)
		return nil, errors.WithMessage(err, "failed to find transactions")
	}
	return trs, nil
}

func (r *repository) FindDepositHistories(ctx context.Context, start, end time.Time) (models.BittrexDepositHistorySlice, error) {
	s := start.Format(timeFormat)
	e := end.Format(timeFormat)
	dhs, err := models.BittrexDepositHistories(
		qm.Where("timestamp >= ? AND timestamp < ?", s, e),
		qm.OrderBy("timestamp ASC"),
	).All(ctx, r.db)
	if err != nil {
		log.Println("failed to find deposit:", "start:", s, "end:", e)
		return nil, errors.WithMessage(err, "failed to find deposit")
	}
	return dhs, nil
}

func (r *repository) FindWithdrawHistories(ctx context.Context, start, end time.Time) (models.BittrexWithdrawHistorySlice, error) {
	s := start.Format(timeFormat)
	e := end.Format(timeFormat)
	whs, err := models.BittrexWithdrawHistories(
		qm.Where("timestamp >= ? AND timestamp < ?", s, e),
		qm.OrderBy("timestamp ASC"),
	).All(ctx, r.db)
	if err != nil {
		log.Println("failed to find withdraw:", "start:", s, "end:", e)
		return nil, errors.WithMessage(err, "failed to find withdraw")
	}
	return whs, nil
}
