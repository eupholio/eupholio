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

package bitflyer

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"time"

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
		n, err := models.BFTransactions().DeleteAll(ctx, db)
		if err != nil {
			return err
		}
		if n > 0 {
			log.Println(n, "records deleted from", models.TableNames.BFTransactions)
		}
	}

	hrs, err := Extract(reader)
	if err != nil {
		return err
	}
	err = repository.CreateTransactions(ctx, hrs.Transactions)
	if err != nil {
		return err
	}
	return nil
}

// ValidateColumnNames checks columns
func ValidateColumnNames(lang string, names []string) error {
	remains := make(map[string]struct{})
	cnSet := columnNamesSet[lang]
	for k := range cnSet {
		remains[k] = struct{}{}
	}
	for _, n := range names {
		if _, ok := cnSet[n]; !ok {
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

type TransactionHistory struct {
	Transactions models.BFTransactionSlice
}

func ExtractFromFile(path string) (*TransactionHistory, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	return Extract(f)
}

// Extract extracts transactions from a reader
func Extract(reader io.Reader) (*TransactionHistory, error) {
	rows, err := extractRecords(NewReader(reader))
	if err != nil {
		return nil, err
	}

	var trs models.BFTransactionSlice

	for _, row := range rows {
		trDate, err := time.Parse("2006/01/02 15:04:05 MST", row.Get(TrDate)+" JST")
		if err != nil {
			return nil, err
		}
		t := row.TrType()
		if t == TrTypeUnknown {
			return nil, fmt.Errorf("unknown type: %v", row)
		}
		tr := &models.BFTransaction{
			TRDate:            trDate,
			Currency:          row.Get(Currency),
			TRType:            t,
			TRPrice:           types.NewDecimal(row.TrPrice()),
			Currency1:         row.Get(Currency1),
			Currency1Quantity: types.NewDecimal(row.Currency1Quantity()),
			Fee:               types.NewDecimal(row.Fee()),
			Currency1JpyRate:  types.NewNullDecimal(row.Currency1JPYRate()),
			Currency2:         null.NewString(row.Currency2(), len(row.Currency2()) > 0),
			Currency2Quantity: types.NewDecimal(row.Currency2Quantity()),
			DealType:          null.NewInt(row.DealType(), true),
			OrderID:           row.OrderID(),
			Remarks:           null.NewString(row.Remarks(), true),
		}
		trs = append(trs, tr)
	}

	trhistory := &TransactionHistory{
		Transactions: trs,
	}
	return trhistory, nil
}

func extractRecords(reader io.Reader) ([]Record, error) {
	r := csv.NewReader(reader)

	csvHead, err := r.Read()
	if err != nil {
		return nil, err
	}

	lang := ""
	for _, l := range []string{En, Jp} {
		err = ValidateColumnNames(l, csvHead)
		if err == nil {
			lang = l
			break
		}
	}

	if err != nil {
		return nil, err
	}

	csvRows, err := r.ReadAll()
	if err != nil {
		return nil, err
	}

	rows, err := MakeRecords(lang, csvHead, csvRows)
	if err != nil {
		return nil, err
	}
	return rows, nil
}
