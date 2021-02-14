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

package coingecko

import (
	"context"
	"fmt"
	"time"

	"github.com/eupholio/eupholio/pkg/httputil"
)

var Currencies = map[int]string{
	1:     "btc",
	279:   "eth",
	44:    "xrp",
	325:   "usdt",
}

var BaseCurrencies = []string{
	"usd",
	"jpy",
}

func DownloadHistoricalPrice(ctx context.Context, currencyCode int, baseCurrency string) ([]byte, error) {
	url := fmt.Sprintf("https://www.coingecko.com/price_charts/export/%d/%s.csv", currencyCode, baseCurrency)
	bs, err := httputil.HttpGet(ctx, url, time.Minute)
	return bs, err
}
