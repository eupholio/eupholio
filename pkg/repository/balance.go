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

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/eupholio/eupholio/models"
)

// Balance

func (r *repository) CreateBalances(ctx context.Context, balances models.BalanceSlice) error {
	for _, balance := range balances {
		err := balance.Insert(ctx, r.ContextExecutor, boil.Infer())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) FindBalanceByCurrencyAndYear(ctx context.Context, currency string, year int) (*models.Balance, error) {
	return models.Balances(
		qm.Where("currency = ? AND year = ?", currency, year),
		qm.Limit(1),
	).One(ctx, r.ContextExecutor)
}

func (r *repository) FindBalancesByYear(ctx context.Context, year int) (models.BalanceSlice, error) {
	bs, err := models.Balances(
		qm.Where("year = ?", year),
	).All(ctx, r.ContextExecutor)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return bs, err
}
