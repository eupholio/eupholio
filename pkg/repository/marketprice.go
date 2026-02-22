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
	"fmt"
	"log"
	"sort"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/eupholio/eupholio/models"
)

// Market Price

func (r *repository) FindLatestMarketPriceByCurrency(ctx context.Context, currency string) (*models.MarketPrice, error) {
	price, err := models.MarketPrices(
		qm.Where("base_currency = ? AND currency = ?", r.baseCurrency, currency),
		qm.OrderBy("time DESC"),
		qm.Limit(1),
	).One(ctx, r.ContextExecutor)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no market price found for %s/%s", currency, r.baseCurrency)
	}
	if price.BaseCurrency != string(r.baseCurrency) {
		return nil, fmt.Errorf("found market price but base currency is invalid")
	}
	return price, err
}

func (r *repository) FindMarketPriceByCurrencyAndTime(ctx context.Context, currency string, tm time.Time) (*models.MarketPrice, error) {
	from := tm.Format(timeFormat)
	to := tm.Add(time.Hour * 48).Format(timeFormat)
	price, err := models.MarketPrices(
		qm.Where("base_currency = ? AND currency = ? AND time >= ? AND time < ?", r.baseCurrency, currency, from, to),
		qm.OrderBy("time ASC"),
		qm.Limit(1),
	).One(ctx, r.ContextExecutor)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no market price found for %s/%s at %s", currency, r.baseCurrency, from)
	}
	if price.BaseCurrency != string(r.baseCurrency) {
		return nil, fmt.Errorf("found market price but base currency is invalid")
	}
	return price, err
}

func (r *repository) CreateMarketPrices(ctx context.Context, marketPrices models.MarketPriceSlice) error {
	for _, marketPrice := range marketPrices {
		err := marketPrice.Insert(ctx, r.ContextExecutor, boil.Infer())
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *repository) AppendMarketPrices(ctx context.Context, marketPrices models.MarketPriceSlice) error {
	if len(marketPrices) == 0 {
		return nil
	}
	sort.SliceStable(marketPrices, func(i, j int) bool {
		return marketPrices[i].Time.Before(marketPrices[j].Time)
	})
	latest, err := r.FindLatestMarketPriceByCurrency(ctx, marketPrices[0].Currency)
	if err != nil {
		latest = nil
	}
	if latest != nil {
		index := sort.Search(len(marketPrices), func(i int) bool {
			return marketPrices[i].Time.After(latest.Time)
		})
		if index < len(marketPrices) {
			log.Printf("append from %s", marketPrices[index].Time.String())
			marketPrices = marketPrices[index:]
		} else {
			log.Printf("up to date %s", latest.Time.String())
			return nil
		}
	}
	return r.CreateMarketPrices(ctx, marketPrices)
}
