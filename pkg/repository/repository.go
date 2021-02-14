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
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/eupholio"
)

func New(db boil.ContextExecutor, baseCurrency currency.Symbol) eupholio.Repository {
	return &repository{
		ContextExecutor: db,
		baseCurrency:    baseCurrency,
	}
}

type repository struct {
	boil.ContextExecutor
	baseCurrency    currency.Symbol
}
