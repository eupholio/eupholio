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

package coincheck

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"

	null "github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/eupholio"
)

// Extractor for BitFlyer
type Extractor struct {
}

// NewExecutor create an executor for BitFlyer
func NewExecutor() *Extractor {
	return &Extractor{}
}

// Execute performs ETL
func (e *Extractor) Execute(ctx context.Context, db boil.ContextExecutor, reader io.Reader, options ...eupholio.Option) error {
	config := &eupholio.Config{}
	for _, o := range options {
		o(config)
	}

	repository := NewRepository(db)

	if config.Overwrite {
		n, err := models.CoincheckHistories().DeleteAll(ctx, db)
		if err != nil {
			return err
		}
		if n > 0 {
			log.Println(n, "records deleted from", models.TableNames.CoincheckHistory)
		}
	}

	hrs, err := Extract(reader)
	if err != nil {
		return err
	}
	err = repository.CreateHistories(ctx, hrs.Entries)
	if err != nil {
		return err
	}
	return nil
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

type History struct {
	Entries models.CoincheckHistorySlice
}

func ExtractFromFile(path string) (*History, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return Extract(f)
}

// Extract extracts transactions from a reader
func Extract(reader io.Reader) (*History, error) {
	rows, err := extractRecords(reader)
	if err != nil {
		return nil, err
	}

	var hs models.CoincheckHistorySlice

	for _, row := range rows {
		tm, err := row.Time()
		if err != nil {
			return nil, err
		}
		op := row.Operation()
		oc := null.NewString(row.OriginalCurrency(), len(row.OriginalCurrency()) > 0)
		h := &models.CoincheckHistory{
			IDCode:           row.ID(),
			Time:             tm,
			Operation:        op,
			Amount:           types.NewDecimal(row.Amount()),
			TradingCurrency:  row.TradingCurrency(),
			Price:            types.NewNullDecimal(row.Price()),
			OriginalCurrency: oc,
			Fee:              types.NewNullDecimal(row.Fee()),
			Comment:          row.Comment(),
		}
		hs = append(hs, h)
	}

	trhistory := &History{
		Entries: hs,
	}
	return trhistory, nil
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
