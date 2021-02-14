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

package etlcmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/eupholio/eupholio/pkg/costmethod"
	"github.com/eupholio/eupholio/pkg/costmethod/mam"
	"github.com/eupholio/eupholio/pkg/costmethod/wam"
	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/repository"
)

const (
	CostMethodWeightedAverage = "wam"
	CostMethodMovingAverage   = "mam"
)

var CostMethods = []string{
	CostMethodWeightedAverage,
	CostMethodMovingAverage,
}

// Calculate updates balance
func Calculate(ctx context.Context, tx *sql.Tx, year int, fiatCurrency currency.Symbol, loc *time.Location, method string, options ...costmethod.Option) error {
	var years []int
	if year == 0 {
		now := time.Now()
		for i := 2008; i <= now.Year(); i++ {
			years = append(years, i)
		}
	} else {
		years = append(years, year)
	}

	if err := currency.InitSymbols(ctx, tx); err != nil {
		return err
	}

	repo := repository.New(tx, fiatCurrency)

	calcs := map[string]costmethod.Calculator{
		CostMethodWeightedAverage: wam.NewCalculator(),
		CostMethodMovingAverage:   mam.NewCalculator(),
	}

	for _, y := range years {
		err := costmethod.CalculateFiatPrice(ctx, repo, y, loc, fiatCurrency)
		if err != nil {
			return err
		}
		config, err := repo.FindConfigByYear(ctx, y)
		if err != nil {
			return err
		}
		m := config.CostMethod
		if method != "" {
			m = method
		}

		log.Printf("calculate %d using %s", y, m)
		calc, ok := calcs[m]
		if !ok {
			return fmt.Errorf("no cost calcuration method found")
		}
		err = costmethod.UpdateBalanceByYear(ctx, repo, y, loc, fiatCurrency, calc, options...)
		if err != nil {
			return err
		}
	}
	return nil
}
