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

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/eupholio"
)

// Config

func (r *repository) FindConfigByYear(ctx context.Context, year int) (*models.Config, error) {
	c, err := models.Configs(
		qm.Where("id = 0 AND year <= ?", year),
		qm.Limit(1),
	).One(ctx, r.ContextExecutor)
	if err == sql.ErrNoRows {
		return eupholio.NewDefaultConfig(year), nil
	} else if err != nil {
		return nil, err
	}
	return c, err
}
