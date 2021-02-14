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
	"log"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/currency"
)

// Event

func (r *repository) CreateEvents(ctx context.Context, events models.EventSlice) error {
	for _, event := range events {
		if event.Quantity.Sign() == 0 {
			continue
		}
		err := event.Insert(ctx, r.ContextExecutor, boil.Infer())
		if err != nil {
			log.Println("failed to insert", event)
			return err
		}
	}
	return nil
}

func (r *repository) FindEvents(ctx context.Context) (models.EventSlice, error) {
	es, err := models.Events(
		qm.OrderBy("time ASC, id ASC"),
	).All(ctx, r.ContextExecutor)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return es, err
}

func (r *repository) FindEventsByStartAndEnd(ctx context.Context, start, end time.Time) (models.EventSlice, error) {
	s := start.UTC().Format(timeFormat)
	e := end.UTC().Format(timeFormat)
	es, err := models.Events(
		qm.Where("time >= ? AND time < ?", s, e),
		qm.OrderBy("time ASC, id ASC"),
	).All(ctx, r.ContextExecutor)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return es, err
}

func (r *repository) FindEventsByCurrencyStartAndEnd(ctx context.Context, cur currency.Symbol, start, end time.Time) (models.EventSlice, error) {
	es, err := models.Events(
		qm.InnerJoin("transactions ON events.transaction_id = transactions.id"),
		qm.Where("currency = ? AND events.time >= ? AND events.time < ?", cur, start, end),
		qm.OrderBy("time ASC, id ASC"),
	).All(ctx, r.ContextExecutor)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return es, err
}

func (r *repository) FindEventsByYear(ctx context.Context, year int, loc *time.Location) (models.EventSlice, error) {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, loc).UTC()
	end := time.Date(year+1, time.January, 1, 0, 0, 0, 0, loc).UTC()
	return r.FindEventsByStartAndEnd(ctx, start, end)
}
