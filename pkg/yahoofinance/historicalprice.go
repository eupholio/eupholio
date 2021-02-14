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

package yahoofinance

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

const DataSourceCode = "yahoofinance"

const (
	Date  = 0
	Close = 4
)

var historicalPriceHeaderColumns = []string{
	"Date", "Open", "High", "Low", "Close", "Adj Close", "Volume",
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
	quote := currency.Symbol(l.currency)
	base := currency.Symbol(l.baseCurrency)
	repo := repository.New(db, base)
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
		var closeTime time.Time
		if quote.IsFiat() && base.IsFiat() {
			closeTime, err = time.Parse("2006-01-02 15:04:05 MST", r[Date]+" 16:59:59 EST")
			if err != nil {
				return err
			}
		} else {
			closeTime, err = time.Parse("2006-01-02 15:04:05 MST", r[Date]+" 23:59:56 UTC")
			if err != nil {
				return err
			}
		}
		close := r[Close]
		if close == "null" {
			continue
		}
		price, ok := new(decimal.Big).SetString(close)
		if !ok {
			return fmt.Errorf("invalid price: %s", close)
		}
		marketPrice := &models.MarketPrice{
			Source:       DataSourceCode,
			Currency:     l.currency,
			Time:         closeTime.UTC(),
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
		return fmt.Errorf("invalid header: expected %d, but got %d: %v", len(historicalPriceHeaderColumns), len(header), header)
	}
	for i, h := range header {
		if h != historicalPriceHeaderColumns[i] {
			return fmt.Errorf("invalid column: %s (%s is expected)", h, historicalPriceHeaderColumns[i])
		}
	}
	return nil
}
