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

package cryptodatadownload

import (
	"bufio"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/repository"
)

const (
	UnixTimestamp = 0
	Date          = 1
	Close         = 6
)

const NumColumns = 9

var historicalPriceHeaderColumns = []string{
	"Unix Timestamp", "Date", "Symbol", "Open", "High", "Low", "Close",
}

type HistoricalPriceLoader struct {
	exchange     string
	currency     string
	baseCurrency string
}

func NewHistoricalPriceLoader(exchange, currency, baseCurrency string) *HistoricalPriceLoader {
	return &HistoricalPriceLoader{
		exchange:     exchange,
		currency:     currency,
		baseCurrency: baseCurrency,
	}
}

func (l *HistoricalPriceLoader) Execute(ctx context.Context, db boil.ContextExecutor, reader io.Reader) error {
	dataSourceCode := "cdd." + l.exchange
	repo := repository.New(db, currency.Symbol(l.baseCurrency))

	if len(l.currency) == 0 || len(l.baseCurrency) == 0 {
		return fmt.Errorf("no symbol")
	}
	bufReader := bufio.NewReaderSize(reader, 4096)
	_, err := bufReader.ReadString('\n')
	if err != nil {
		return err
	}
	r := csv.NewReader(bufReader)
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
		marketPrice, err := l.processRecord(r, dataSourceCode)
		if err != nil {
			return err
		}
		if marketPrice != nil {
			marketPrices = append(marketPrices, marketPrice)
		}
	}

	err = repo.AppendMarketPrices(ctx, marketPrices)
	if err != nil {
		return err
	}
	return nil
}

var timestampRe = regexp.MustCompile(`(\d+)\.(\d+)`)

func parseTimestamp(s string) (time.Time, error) {
	m := timestampRe.FindAllStringSubmatch(s, -1)
	if len(m) == 0 {
		return time.Time{}, fmt.Errorf("invalid format %s", s)
	}
	m0 := m[0]
	if len(m0) < 2 {
		return time.Time{}, fmt.Errorf("invalid format %s", s)
	}
	unixtimeSec, err := strconv.Atoi(m0[1])
	if err != nil {
		return time.Time{}, err
	}
	date := time.Unix(int64(unixtimeSec), 0).UTC()
	return date, nil
}

func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02 03-PM", s)
}

func (l *HistoricalPriceLoader) processRecord(r []string, dataSourceCode string) (*models.MarketPrice, error) {
	if len(r) != NumColumns {
		return nil, fmt.Errorf("invalid record")
	}
	_, err := parseTimestamp(r[UnixTimestamp])
	if err != nil {
		return nil, err
	}
	date, err := parseDate(r[Date])
	if err != nil {
		return nil, err
	}
	closeTime := date.Add(time.Hour - time.Second)
	close := r[Close]
	price, ok := new(decimal.Big).SetString(close)
	if !ok {
		return nil, fmt.Errorf("invalid price: %s", close)
	}

	marketPrice := &models.MarketPrice{
		Source:       dataSourceCode,
		Currency:     l.currency,
		Time:         closeTime,
		BaseCurrency: l.baseCurrency,
		Price:        types.NewDecimal(price),
	}
	return marketPrice, nil
}

func (l *HistoricalPriceLoader) validateHeader(header []string) error {
	if len(header) != NumColumns {
		return fmt.Errorf("invalid header: expected %d, but got %d: %v", NumColumns, len(header), header)
	}
	for i, h := range historicalPriceHeaderColumns {
		if h != header[i] {
			return fmt.Errorf("invalid column: %s (%s is expected)", header[i], h)
		}
	}
	return nil
}
