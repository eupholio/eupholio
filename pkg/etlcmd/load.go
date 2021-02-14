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

package etlcmd

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/eupholio/eupholio/pkg/coingecko"
	"github.com/eupholio/eupholio/pkg/cryptodatadownload"
	"github.com/eupholio/eupholio/pkg/eupholio"
	"github.com/eupholio/eupholio/pkg/yahoofinance"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

var coingechoHistoricalPriceFilenameRE = regexp.MustCompile("([a-z]+)-([a-z]+)-max.csv")

func LoadCoingeckoHistoricalPrice(db boil.ContextExecutor, args []string) error {
	ctx := context.Background()
	for _, arg := range args {
		filename := filepath.Base(arg)
		submatches := coingechoHistoricalPriceFilenameRE.FindAllStringSubmatch(filename, -1)
		if len(submatches) != 1 {
			return fmt.Errorf("invalid filename: %s", filename)
		}
		currency := strings.ToUpper(submatches[0][1])
		baseCurrency := strings.ToUpper(submatches[0][2])
		loader := coingecko.NewHistoricalPriceLoader(currency, baseCurrency)
		if err := load(ctx, db, arg, loader); err != nil {
			return err
		}
	}
	return nil
}

var yahooFinanceHistoricalPriceFilenameRE = regexp.MustCompile("([A-Z]+)-([A-Z]+).csv")

func LoadYahooFinanceHistoricalPrice(db *sql.DB, args []string) error {
	ctx := context.Background()
	for _, arg := range args {
		filename := filepath.Base(arg)
		submatches := yahooFinanceHistoricalPriceFilenameRE.FindAllStringSubmatch(filename, -1)
		if len(submatches) != 1 {
			return fmt.Errorf("invalid filename: %s", filename)
		}
		currency := strings.ToUpper(submatches[0][1])
		baseCurrency := strings.ToUpper(submatches[0][2])
		log.Println("loading", currency, "/", baseCurrency)
		loader := yahoofinance.NewHistoricalPriceLoader(currency, baseCurrency)
		err := load(ctx, db, arg, loader)
		if err != nil {
			return err
		}
	}
	return nil
}

var cddHistoricalPriceFilenameRE = regexp.MustCompile("(Bittrex|Poloniex)_([A-Z]+)(USD|BTC|ETH)_1h.csv")

func LoadCDDHistoricalPrice(db *sql.DB, args []string) error {
	ctx := context.Background()
	for _, arg := range args {
		filename := filepath.Base(arg)
		submatches := cddHistoricalPriceFilenameRE.FindAllStringSubmatch(filename, -1)
		if len(submatches) != 1 {
			return fmt.Errorf("invalid filename: %s", filename)
		}
		exchange := strings.ToLower(submatches[0][1])
		currency := strings.ToUpper(submatches[0][2])
		baseCurrency := strings.ToUpper(submatches[0][3])
		log.Println("loading", currency, "/", baseCurrency)
		loader := cryptodatadownload.NewHistoricalPriceLoader(exchange, currency, baseCurrency)
		err := load(ctx, db, arg, loader)
		if err != nil {
			return err
		}
	}
	return nil
}

func load(ctx context.Context, db boil.ContextExecutor, path string, extractor eupholio.Loader) error {
	reader, err := os.Open(path)
	if err != nil {
		return err
	}
	defer reader.Close()
	err = extractor.Execute(ctx, db, reader)
	if err != nil {
		return err
	}
	return nil
}
