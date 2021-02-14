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
)

// Entry

func (r *repository) CreateEntries(ctx context.Context, entries models.EntrySlice) error {
	for _, entry := range entries {
		err := entry.Insert(ctx, r.ContextExecutor, boil.Infer())
		if err != nil {
			log.Println("failed to insert", entry)
			return err
		}
	}
	return nil
}

func (r *repository) UpdateEntries(ctx context.Context, entries models.EntrySlice) error {
	for _, entry := range entries {
		_, err := entry.Update(ctx, r.ContextExecutor, boil.Infer())
		if err != nil {
			log.Println("failed to insert", entry)
			return err
		}
	}
	return nil
}

func (r *repository) FindEntriesByStartAndEnd(ctx context.Context, start, end time.Time) (models.EntrySlice, error) {
	s := start.UTC().Format(timeFormat)
	e := end.UTC().Format(timeFormat)
	es, err := models.Entries(
		qm.Where("time >= ? AND time < ?", s, e),
		qm.OrderBy("time ASC, id ASC"),
	).All(ctx, r.ContextExecutor)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return es, err
}

func (r *repository) FindEntriesByYear(ctx context.Context, year int, loc *time.Location) (models.EntrySlice, error) {
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, loc).UTC()
	end := time.Date(year+1, time.January, 1, 0, 0, 0, 0, loc).UTC()
	return r.FindEntriesByStartAndEnd(ctx, start, end)
}
