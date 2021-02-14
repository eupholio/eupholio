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
	"log"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/eupholio/eupholio/models"
)

const timeFormat = "2006/01/02 15:04:05"

type Repository interface {
	FindHistories(ctx context.Context, start, end time.Time) (models.CoincheckHistorySlice, error)
	FindHistoriesByYear(ctx context.Context, year int, location *time.Location) (models.CoincheckHistorySlice, error)
	CreateHistories(ctx context.Context, hs models.CoincheckHistorySlice) error
}

func NewRepository(db boil.ContextExecutor) Repository {
	return &repository{
		db: db,
	}
}

type repository struct {
	db boil.ContextExecutor
}

func (r *repository) FindHistories(ctx context.Context, start time.Time, end time.Time) (models.CoincheckHistorySlice, error) {
	s := start.UTC().Format(timeFormat)
	e := end.UTC().Format(timeFormat)
	q := models.CoincheckHistories(
		qm.Where("time >= ? AND time < ?", s, e),
		qm.OrderBy("time ASC"),
	)
	trs, err := q.All(ctx, r.db)
	if err != nil {
		log.Println("failed to find history:", "start:", s, "end:", e)
		return nil, errors.WithMessage(err, "failed to find history")
	}
	return trs, err
}

func (r *repository) FindHistoriesByYear(ctx context.Context, year int, loc *time.Location) (models.CoincheckHistorySlice, error) {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, loc)
	end := time.Date(year+1, time.January, 1, 0, 0, 0, 0, loc)
	return r.FindHistories(ctx, start, end)
}

func (r *repository) CreateHistories(ctx context.Context, trs models.CoincheckHistorySlice) error {
	for _, tr := range trs {
		err := tr.Insert(ctx, r.db, boil.Infer())
		if err != nil {
			return err
		}
	}
	return nil
}
