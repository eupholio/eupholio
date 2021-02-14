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

package yahoofinance

import (
	"context"
	"fmt"
	"time"

	"github.com/eupholio/eupholio/pkg/httputil"
)

func GetCurrencySymbol(name string) string {
	currencies := map[string]string{}
	if s, ok := currencies[name]; ok {
		return s
	}
	return name
}

func DownloadHistoricalPrice(ctx context.Context, baseCurrency string, quote string) ([]byte, error) {
	period1 := 1199145600 // 2008-01-01
	period2 := time.Now().Unix()
	quoteCurrency := GetCurrencySymbol(quote)
	interval := "1d"
	url := fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/download/%s-%s?period1=%d&period2=%d&interval=%s&events=history&includeAdjustedClose=true", quoteCurrency, baseCurrency, period1, period2, interval)
	switch baseCurrency {
	case "JPY":
		switch quote {
		case "USD":
			url = fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/download/%s=X?period1=%d&period2=%d&interval=%s&events=history&includeAdjustedClose=true", baseCurrency, period1, period2, interval)
		case "EUR":
			code := "EURJPY"
			url = fmt.Sprintf("https://query1.finance.yahoo.com/v7/finance/download/%s=X?period1=%d&period2=%d&interval=%s&events=history&includeAdjustedClose=true", code, period1, period2, interval)
		}
	}
	bs, err := httputil.HttpGet(ctx, url, time.Minute)
	if err != nil {
		return nil, err
	}
	return bs, nil
}
