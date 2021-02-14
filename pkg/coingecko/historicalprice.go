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

package coingecko

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/repository"
)

// DataSourceCode is the data source name
const DataSourceCode = "coingecko"

var historicalPriceHeaderColumns = []string{
	"snapped_at",
	"price",
	"market_cap",
	"total_volume",
}

type HistoricalPriceLoader struct {
	currency     string
	baseCurrency string
}

func NewHistoricalPriceLoader(currency, baseCurrency string) *HistoricalPriceLoader {
	return &HistoricalPriceLoader{
		currency:     currency,
		baseCurrency: baseCurrency,
	}
}

func (l *HistoricalPriceLoader) Execute(ctx context.Context, db boil.ContextExecutor, reader io.Reader) error {
	repo := repository.New(db, currency.Symbol(l.baseCurrency))
	if len(l.currency) == 0 || len(l.baseCurrency) == 0 {
		return fmt.Errorf("no symbol")
	}
	r := csv.NewReader(reader)
	header, err := r.Read()
	if err != nil {
		return err
	}
	if err := l.validateHeader(header); err != nil {
		return err
	}
	records, err := r.ReadAll()
	if err != nil {
		return err
	}

	var marketPrices models.MarketPriceSlice
	for _, r := range records {
		if len(r) != len(historicalPriceHeaderColumns) {
			return fmt.Errorf("invalid record")
		}
		snappedAt, err := time.Parse("2006-01-02 15:04:05 MST", r[0])
		if err != nil {
			return err
		}
		price, ok := new(decimal.Big).SetString(r[1])
		if !ok {
			return fmt.Errorf("invalid price: %s", r[1])
		}
		marketPrice := &models.MarketPrice{
			Source:       DataSourceCode,
			Currency:     l.currency,
			Time:         snappedAt,
			BaseCurrency: l.baseCurrency,
			Price:        types.NewDecimal(price),
		}
		marketPrices = append(marketPrices, marketPrice)
	}

	err = repo.AppendMarketPrices(ctx, marketPrices)
	if err != nil {
		return err
	}
	return nil
}

func (l *HistoricalPriceLoader) validateHeader(header []string) error {
	if len(historicalPriceHeaderColumns) != len(header) {
		return fmt.Errorf("invalid header: expected %d, but got %d", len(historicalPriceHeaderColumns), len(header))
	}
	for i, h := range header {
		if h != historicalPriceHeaderColumns[i] {
			return fmt.Errorf("invalid column: %s (%s is expected)", h, historicalPriceHeaderColumns[i])
		}
	}
	return nil
}
