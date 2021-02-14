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

const recordTimeFormat = "2006/1/2 15:04:05"

type ErrorWithLineInfo struct {
	line   int
	column string
	err    error
}

func (e *ErrorWithLineInfo) Error() string {
	if len(e.column) == 0 {
		return fmt.Sprintf("line %d: %s", e.line, e.err.Error())
	}
	return fmt.Sprintf("line %d(%s): %s", e.line, e.column, e.err.Error())
}

func errorWithLineNumber(n int, err error) error {
	return &ErrorWithLineInfo{
		line: n,
		err:  err,
	}
}

func errorWithLineNumberAndColumn(n int, column string, err error) error {
	return &ErrorWithLineInfo{
		line:   n,
		column: column,
		err:    err,
	}
}

// Extractor for Poloniex trades
type Extractor struct {
	loc *time.Location
}

// NewExtractor create an executor for Poloniex trades
func NewExtractor(loc *time.Location) *Extractor {
	return &Extractor{
		loc: loc,
	}
}

// Execute performs ETL
func (e *Extractor) Execute(ctx context.Context, db boil.ContextExecutor, reader io.Reader, options ...eupholio.Option) error {
	config := &eupholio.Config{}
	for _, o := range options {
		o(config)
	}

	trades, err := e.extract(reader)
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
func (e *Extractor) extract(reader io.Reader) (models.CryptactCustomSlice, error) {
	rows, err := extractRecords(reader)
	if err != nil {
		return nil, errorWithLineNumber(1, err)
	}

	var customs models.CryptactCustomSlice

	for i, row := range rows {
		n := i + 2
		timestamp, err := row.Timestamp(e.loc)
		if err != nil {
			return nil, errorWithLineNumber(n, err)
		}
		volume, err := row.GetAsDecimal(VolumeColumn)
		if err != nil {
			return nil, errorWithLineNumber(n, err)
		}
		price, err := row.GetAsNullDecimal(PriceColumn)
		if err != nil {
			return nil, errorWithLineNumber(n, err)
		}
		fee, err := row.GetAsDecimal(FeeColumn)
		if err != nil {
			return nil, errorWithLineNumber(n, err)
		}
		custom := &models.CryptactCustom{
			Timestamp: timestamp,
			Action:    row.Get(ActionColumn),
			Source:    row.Get(SourceColumn),
			Base:      row.Get(BaseColumn),
			Volume:    types.NewDecimal(volume),
			Price:     types.NewNullDecimal(price),
			Counter:   row.Get(CounterColumn),
			Fee:       types.NewDecimal(fee),
			FeeCcy:    row.Get(FeeCcyColumn),
		}
		customs = append(customs, custom)
	}

	return customs, nil
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

func (r Record) Timestamp(loc *time.Location) (time.Time, error) {
	date := r.Get(TimestampColumn)
	return time.ParseInLocation(recordTimeFormat, date, loc)
}

func (r Record) GetAsDecimal(name string) (*decimal.Big, error) {
	s := r.Get(name)
	if b, ok := new(decimal.Big).SetString(s); ok {
		return b, nil
	} else {
		return nil, fmt.Errorf("invalid decimal '%s' in %s column", s, name)
	}
}

func (r Record) GetAsNullDecimal(name string) (*decimal.Big, error) {
	s := r.Get(name)
	// s = strings.TrimRight(strings.TrimLeft(s, " "), " ")
	if len(s) == 0 {
		return nil, nil
	}
	return r.GetAsDecimal(name)
}
