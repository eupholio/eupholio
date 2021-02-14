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

package main

import (
	"bytes"
	"context"
	"database/sql"
	gosql "database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/eupholio/eupholio/pkg/bitflyer"
	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/etlcmd"
	"github.com/eupholio/eupholio/pkg/querycmd"
	"github.com/eupholio/eupholio/pkg/repository"
	"github.com/eupholio/eupholio/pkg/yahoofinance"
)

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

var db *gosql.DB

func TestMain(m *testing.M) {
	var err error
	db, err = openDB()
	if err != nil {
		panic(err)
	}
	m.Run()
	os.Exit(0)
}

func TestLoad(t *testing.T) {
	ctx := context.Background()
	err := withRollback(ctx, db, func(tx *gosql.Tx) {
		t.Log("loading price")
		testLoad(t, ctx, tx)
		repo := repository.New(tx, "JPY")
		tm := time.Date(2018, time.January, 1, 0, 0, 0, 0, jst)
		marketPrice, err := repo.FindMarketPriceByCurrencyAndTime(ctx, "BTC", tm)
		if err != nil {
			t.Error(err)
		}
		if marketPrice.Price.String() != "1539985.8750000000" {
			t.Error("price not match", marketPrice.Price)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func testLoad(t *testing.T, ctx context.Context, db boil.ContextExecutor) {
	loader := yahoofinance.NewHistoricalPriceLoader("BTC", "JPY")
	buf := bytes.NewBufferString(btcMarketPriceCsv)
	if err := loader.Execute(ctx, db, buf); err != nil {
		t.Fatal(err)
	}
	loader = yahoofinance.NewHistoricalPriceLoader("ETH", "JPY")
	buf = bytes.NewBufferString(ethMarketPriceCsv)
	if err := loader.Execute(ctx, db, buf); err != nil {
		t.Fatal(err)
	}
}

func TestImport(t *testing.T) {
	ctx := context.Background()
	err := withRollback(ctx, db, func(tx *gosql.Tx) {
		testLoad(t, ctx, tx)
		t.Log("import")
		testImportBitflyer(t, ctx, tx)
		testImportBittrex(t, ctx, tx)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func testImportBitflyer(t *testing.T, ctx context.Context, tx *sql.Tx) {
	err := etlcmd.ImportBitflyerData(ctx, tx, []string{"../testdata/TradeHistory.csv"}, true)
	if err != nil {
		t.Fatal(err)
	}
	repository := bitflyer.NewRepository(tx)
	start := time.Date(2015, time.January, 1, 0, 0, 0, 0, jst)
	end := time.Date(2020, time.January, 1, 0, 0, 0, 0, jst)
	ts, err := repository.FindTransactions(ctx, start, end)
	if err != nil {
		t.Error(err)
	}
	if len(ts) != 7 {
		t.Error("import failed")
	}
	//bitflyer.NewTableWriter(os.Stderr).PrintTransactionsShort(ts)
}

func testImportBittrex(t *testing.T, ctx context.Context, tx *sql.Tx) {
	err := etlcmd.ImportBittrexData(ctx, tx, []string{"../testdata/BittrexDeposit.csv"}, true, "deposit")
	if err != nil {
		t.Fatal(err)
	}
	err = etlcmd.ImportBittrexData(ctx, tx, []string{"../testdata/BittrexWithdraw.csv"}, true, "withdraw")
	if err != nil {
		t.Fatal(err)
	}
	err = etlcmd.ImportBittrexData(ctx, tx, []string{"../testdata/BittrexOrderHistory.csv"}, true, "order")
	if err != nil {
		t.Fatal(err)
	}
}

func TestTranslate(t *testing.T) {
	ctx := context.Background()
	err := withRollback(ctx, db, func(tx *gosql.Tx) {
		testLoad(t, ctx, tx)
		testImportBitflyer(t, ctx, tx)
		testImportBittrex(t, ctx, tx)
		t.Log("translate")
		err := etlcmd.Translate(ctx, tx, 0, jst, currency.JPY)
		if err != nil {
			t.Error(err)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCalculate(t *testing.T) {
	ctx := context.Background()
	source := "yahoofinance"
	fiat := currency.JPY
	err := withRollback(ctx, db, func(tx *gosql.Tx) {
		repo := repository.New(tx, fiat)
		testLoad(t, ctx, tx)
		testImportBitflyer(t, ctx, tx)
		testImportBittrex(t, ctx, tx)
		err := etlcmd.Translate(ctx, tx, 0, jst, currency.JPY)
		if err != nil {
			t.Error(err)
		}
		t.Log("calculate")
		for _, cfg := range []struct {
			year   int
			method string
		}{
			{2017, "wam"},
			{2018, "wam"},
			{2019, "wam"},
			{2020, "wam"},
		} {
			year := cfg.year
			method := cfg.method
			events, err := repo.FindEventsByYear(ctx, year, jst)
			if err != nil {
				t.Error(err)
			}
			logAsTable(t, fmt.Sprintf("events of %d", year), events)
			err = etlcmd.Calculate(ctx, tx, year, fiat, jst, method)
			if err != nil {
				t.Error(err)
			}
			buf := bytes.NewBuffer(nil)
			querycmd.QueryTransactions(ctx, buf, tx, year, jst, fiat.String(), source, querycmd.OutputFormatTable)
			t.Log("transactions \n", buf.String())
			bs, err := repo.FindBalancesByYear(ctx, year)
			if err != nil {
				t.Error(err)
			}
			if len(bs) == 0 {
				t.Error("empty result")
			}
			logAsTable(t, fmt.Sprintf("balance of %d", year), bs)
		}
	})
	if err != nil {
		t.Fatal(err)
	}
}

func logAsTable(t *testing.T, msg string, value interface{}) {
	buf := bytes.NewBuffer(nil)
	err := querycmd.NewTableWriter(buf).Write(value)
	if err != nil {
		t.Error(err)
	}
	t.Log(fmt.Sprintf("%s\n%s", msg, buf.String()))
}
