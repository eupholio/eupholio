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

package bittrex

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"time"

	"github.com/ericlagergren/decimal"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/eupholio"
)

// Extractor for Bittrex
type Extractor struct {
}

// NewExtractor create an executor for Bittrex
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Execute performs ETL
func (e *Extractor) Execute(ctx context.Context, db boil.ContextExecutor, reader io.Reader, options ...eupholio.Option) error {
	config := &eupholio.Config{}
	for _, o := range options {
		o(config)
	}

	if config.Overwrite {
		n, err := models.BittrexOrderHistories().DeleteAll(ctx, db)
		if err != nil {
			return err
		}
		if n > 0 {
			log.Println(n, "records deleted from", models.TableNames.BittrexOrderHistory)
		}
	}

	hrs, err := Extract(reader)
	if err != nil {
		return err
	}
	trs := hrs.Orders
	for _, tr := range trs {
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

type OrderHistory struct {
	Orders models.BittrexOrderHistorySlice
}

// Extract extracts transactions from a reader
func Extract(reader io.Reader) (*OrderHistory, error) {
	rows, err := extractRecords(reader)
	if err != nil {
		return nil, err
	}

	var ohs models.BittrexOrderHistorySlice

	for _, row := range rows {
		timestamp, err := row.TimeStamp()
		if err != nil {
			return nil, err
		}
		t := row.OrderType()
		if t == OrderTypeUnknown {
			return nil, err
		}
		closed, err := row.Closed()
		if err != nil {
			return nil, err
		}
		tifID, err := row.TimeInForceTypeID()
		if err != nil {
			return nil, err
		}
		oh := &models.BittrexOrderHistory{
			UUID:              row.UUID(),
			Exchange:          row.Exchange(),
			OrderType:         row.OrderType(),
			Timestamp:         timestamp,
			Limit:             types.NewDecimal(row.Limit()),
			Quantity:          types.NewDecimal(row.Quantity()),
			QuantityRemaining: types.NewDecimal(row.QuantityRemaining()),
			Commission:        types.NewDecimal(row.Commission()),
			Price:             types.NewDecimal(row.Price()),
			PricePerUnit:      types.NewDecimal(row.PricePerUnit()),
			IsConditional:     row.IsConditional(),
			ImmediateOrCancel: row.ImmediateOrCancel(),
			Closed:            closed,
			TimeInForceTypeID: tifID,
			TimeInForce:       null.StringFrom(row.TimeInForce()),
		}
		if row.IsConditional() {
			oh.Condition = null.StringFrom(row.Condition())
			oh.ConditionTarget = types.NewNullDecimal(row.ConditionTarget())
		}
		ohs = append(ohs, oh)
	}

	orderHistory := &OrderHistory{
		Orders: ohs,
	}
	return orderHistory, nil
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

// DepositExtractor loads depoist history
type DepositExtractor struct {
}

func NewDepositExtractor() *DepositExtractor {
	return &DepositExtractor{}
}

func (e *DepositExtractor) Execute(ctx context.Context, db boil.ContextExecutor, reader io.Reader, options ...eupholio.Option) error {
	r := csv.NewReader(reader)
	for {
		date, symbol, quantity, status, err := readDepositOrWithdrawCsv(r)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		depost := &models.BittrexDepositHistory{
			ID:        0,
			Timestamp: date,
			Currency:  symbol,
			Quantity:  types.NewDecimal(quantity),
			Status:    status,
		}
		err = depost.Insert(ctx, db, boil.Infer())
		if err != nil {
			return err
		}
	}
	return nil
}

// WithdrawExtractor loads withdraw history
type WithdrawExtractor struct {
}

func (e *WithdrawExtractor) Execute(ctx context.Context, db boil.ContextExecutor, reader io.Reader, options ...eupholio.Option) error {
	r := csv.NewReader(reader)
	for {
		date, symbol, quantity, status, err := readDepositOrWithdrawCsv(r)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		depost := &models.BittrexWithdrawHistory{
			ID:        0,
			Timestamp: date,
			Currency:  symbol,
			Quantity:  types.NewDecimal(quantity),
			Status:    status,
		}
		err = depost.Insert(ctx, db, boil.Infer())
		if err != nil {
			return err
		}
	}
	return nil
}

func NewWithdrawExtractor() *WithdrawExtractor {
	return &WithdrawExtractor{}
}

func readDepositOrWithdrawCsv(r *csv.Reader) (date time.Time, symbol string, quantity *decimal.Big, status string, err error) {
	timeRowFormat := "2006/01/02 15:04:05"
	if timeRow, e := r.Read(); e != nil {
		err = e
		return
	} else if date, err = time.Parse(timeRowFormat, timeRow[0]); err != nil {
		return
	}

	if symbolRow, e := r.Read(); e != nil {
		err = e
		return
	} else {
		symbol = symbolRow[0]
	}

	var ok bool
	if quantityRow, e := r.Read(); e != nil {
		err = e
		return
	} else if quantity, ok = new(decimal.Big).SetString(quantityRow[0]); !ok {
		err = fmt.Errorf("invalid quantity row: %v", quantityRow)
		return
	}

	if statusRow, e := r.Read(); e != nil {
		err = e
	} else {
		status = statusRow[0]
		if status != "Completed" {
			log.Println("unknown status:", status)
		}
	}
	return
}
