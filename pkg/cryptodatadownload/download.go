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
	"context"
	"fmt"
	"log"
	"time"

	"github.com/eupholio/eupholio/pkg/httputil"
)

// https://www.cryptodatadownload.com/cdd/Poloniex_BTCUSDT_1h.csv
var PoloniexPairs = map[string][]string{
	"USDT": {"BTC", "ETH", "LTC", "XRP", "DASH", "ETC", "BAT", "ZRX", "EOS", "LSK", "REP"},
	"BTC":  {"XRP", "ETH", "LTC", "XMR", "DASH", "LSK", "POT", "NXT", "ETC", "XEM", "DOGE"},
	"ETH":  {"ETC", "OMG", "ZEC", "LSK", "GNO", "REP", "GAS", "ZRX"},
}

// https://www.cryptodatadownload.com/cdd/Bittrex_BTCUSD_1h.csv
var BittrexPairs = map[string][]string{
	"USD": {"BTC", "ETH", "LTC", "NEO", "ETC", "OMG", "XMR", "DASH"},
	"BTC": {"XRP", "ETH", "LTC", "NEO", "ETC", "ZEC", "XLM", "WAVES", "ADA"},
	"ETH": {"XRP", "LTC", "NEO", "ADA", "OMG", "XMR", "ZEC", "XLM", "DASH"},
}
var Exchanges = map[string]map[string][]string{
	"Poloniex": PoloniexPairs,
	"Bittrex":  BittrexPairs,
}

func HistoricalPriceFilename(exchange, baseCurrency, quoteCurrency string) string {
	return fmt.Sprintf("%s_%s%s_1h.csv", exchange, quoteCurrency, baseCurrency)
}

func DownloadHistoricalPrice(ctx context.Context, exchange, baseCurrency, quoteCurrency string) ([]byte, error) {
	filenamePart := HistoricalPriceFilename(exchange, baseCurrency, quoteCurrency)
	url := fmt.Sprintf("https://www.cryptodatadownload.com/cdd/%s", filenamePart)
	bs, err := httputil.HttpGet(ctx, url, time.Minute)
	if err == httputil.ErrNotFound {
		log.Printf("%s not found", url)
	}
	return bs, err
}
