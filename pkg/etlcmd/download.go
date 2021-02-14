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
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/eupholio/eupholio/pkg/coingecko"
	"github.com/eupholio/eupholio/pkg/cryptodatadownload"
	"github.com/eupholio/eupholio/pkg/httputil"
	"github.com/eupholio/eupholio/pkg/yahoofinance"
)

func DownloadCoingeckoHistoricalPrice(dir string) error {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	ctx := context.Background()
	for i, c := range coingecko.Currencies {
		for _, b := range coingecko.BaseCurrencies {
			outputFilepath := filepath.Join(dir, fmt.Sprintf("%s-%s-max.csv", c, b))
			log.Printf("downloading %s", outputFilepath)
			bs, err := coingecko.DownloadHistoricalPrice(ctx, i, b)
			if err != nil {
				return err
			}
			err = ioutil.WriteFile(outputFilepath, bs, 0644)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func DownloadCryptoDataDownloadHistoricalPrice(dir string) error {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	for e, pairs := range cryptodatadownload.Exchanges {
		for b, cs := range pairs {
			for _, c := range cs {
				ctx := context.Background()
				outputFilename := fmt.Sprintf("%s_%s%s_1h.csv", e, c, b)
				outputFilepath := filepath.Join(dir, outputFilename)
				log.Printf("downloading %s", outputFilepath)
				bs, err := cryptodatadownload.DownloadHistoricalPrice(ctx, e, b, c)
				if err != nil {
					if err == httputil.ErrNotFound {
						log.Println(outputFilename, "not found")
						continue
					}
					return err
				}
				err = ioutil.WriteFile(outputFilepath, bs, 0644)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func DownloadYahooFinanceHistoricalPrice(dir string, baseCurrencies []string, symbols []string) error {
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return err
	}

	ctx := context.Background()
	for _, baseCurrency := range baseCurrencies {
		for _, c := range symbols {
			quoteCurrency := yahoofinance.GetCurrencySymbol(c)
			outputFilename := fmt.Sprintf("%s-%s.csv", quoteCurrency, baseCurrency)
			outputFilepath := filepath.Join(dir, outputFilename)
			bs, err := yahoofinance.DownloadHistoricalPrice(ctx, baseCurrency, c)
			if err != nil {
				if err == httputil.ErrNotFound {
					log.Println(outputFilename, "not found")
					continue
				}
				return err
			}
			err = ioutil.WriteFile(outputFilepath, bs, 0644)
			if err != nil {
				return err
			}
		}
	}
	return nil
}
