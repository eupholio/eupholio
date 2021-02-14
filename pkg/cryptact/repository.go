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

package cryptact

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
	FindCustoms(ctx context.Context, start, end time.Time) (models.CryptactCustomSlice, error)
}

func NewRepository(db boil.ContextExecutor) Repository {
	return &repository{
		db: db,
	}
}

type repository struct {
	db boil.ContextExecutor
}

func (r *repository) FindCustoms(ctx context.Context, start, end time.Time) (models.CryptactCustomSlice, error) {
	s := start.Format(timeFormat)
	e := end.Format(timeFormat)
	ts, err := models.CryptactCustoms(
		qm.Where("timestamp >= ? AND timestamp < ?", s, e),
		qm.OrderBy("timestamp ASC"),
	).All(ctx, r.db)
	if err != nil {
		log.Println("failed to find customs:", "start:", s, "end:", e)
		return nil, errors.WithMessage(err, "failed to find custons")
	}
	return ts, nil
}
