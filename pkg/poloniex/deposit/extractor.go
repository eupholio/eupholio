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

package deposit

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/eupholio"
)

const recordTimeFormat = "2006-01-02 15:04:05"

// Extractor for Poloniex trades
type Extractor struct {
}

// NewExtractor create an executor for Poloniex trades
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Execute performs ETL
func (e *Extractor) Execute(ctx context.Context, db boil.ContextExecutor, reader io.Reader, options ...eupholio.Option) error {
	config := &eupholio.Config{}
	for _, o := range options {
		o(config)
	}

	trades, err := extract(reader)
	if err != nil {
		return err
	}
	for _, tr := range trades {
		err := tr.Insert(ctx, db, boil.Infer())
		if err != nil {
			if config.Debug {
				log.Print(tr)
			}
			return err
		}
	}

	return nil
}

// extract extracts transactions from a reader
func extract(reader io.Reader) (models.PoloniexDepositSlice, error) {
	rows, err := extractRecords(reader)
	if err != nil {
		return nil, err
	}

	var deposits models.PoloniexDepositSlice

	for _, row := range rows {
		date, err := row.Date()
		if err != nil {
			return nil, err
		}
		amount, err := row.GetAsDecimal(AmountColumn)
		if err != nil {
			return nil, err
		}
		deposit := &models.PoloniexDeposit{
			ID:       0,
			Date:     date,
			Currency: row.Get(CurrencyColumn),
			Amount:   types.NewDecimal(amount),
			Address:  row.Get(AddressColumn),
			Status:   row.Get(StatusColumn),
		}
		deposits = append(deposits, deposit)
	}

	return deposits, nil
}

func extractRecords(reader io.Reader) ([]Record, error) {
	r := csv.NewReader(reader)

	csvHead, err := r.Read()
	if err != nil {
		return nil, err
	}

	if err := ValidateColumnNames(csvHead); err != nil {
		return nil, err
	}

	csvRows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	rows, err := MakeRecords(csvHead, csvRows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// ValidateColumnNames checks columns
func ValidateColumnNames(names []string) error {
	remains := make(map[string]struct{})
	for k := range columnNamesSet {
		remains[k] = struct{}{}
	}
	for _, n := range names {
		if _, ok := columnNamesSet[n]; !ok {
			return fmt.Errorf("unknown column %s found", n)
		}
		delete(remains, n)
	}
	if len(remains) > 0 {
		for k := range remains {
			return fmt.Errorf("column %s not found", k)
		}
	}
	return nil
}

func MakeRecords(head []string, rows [][]string) ([]Record, error) {
	records := make([]Record, 0, len(rows))
	for _, row := range rows {
		record := make(Record)
		for i, col := range head {
			record[col] = row[i]
		}
		records = append(records, record)
	}
	return records, nil
}

type Record map[string]string

func (r Record) Get(name string) string {
	return r[name]
}

func (r Record) Date() (time.Time, error) {
	date := r.Get(DateColumn)
	return time.ParseInLocation(recordTimeFormat, date, time.UTC)
}

func (r Record) GetAsDecimal(name string) (*decimal.Big, error) {
	s := r.Get(name)
	if b, ok := new(decimal.Big).SetString(s); ok {
		return b, nil
	} else {
		return nil, fmt.Errorf("invalid decimal %s", s)
	}
}
