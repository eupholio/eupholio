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
	"log"
	"time"

	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/bitflyer"
	"github.com/eupholio/eupholio/pkg/bittrex"
	"github.com/eupholio/eupholio/pkg/coincheck"
	"github.com/eupholio/eupholio/pkg/cryptact"
	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/eupholio"
	"github.com/eupholio/eupholio/pkg/poloniex"
	"github.com/eupholio/eupholio/pkg/repository"
)

func Translate(ctx context.Context, tx *sql.Tx, year int, jst *time.Location, fiat currency.Symbol) error {
	translators := map[string]eupholio.Translator{
		"bitflyer":  bitflyer.NewTranslator(),
		"coincheck": coincheck.NewTranslator(),
		"bittrex":   bittrex.NewTranslator(currency.JPY),
		"poloniex":  poloniex.NewTranslator(currency.JPY),
		"cryptact":  cryptact.NewTranslator(currency.JPY),
	}
	start := time.Date(year, time.January, 1, 0, 0, 0, 0, jst)
	end := time.Date(year+1, time.January, 1, 0, 0, 0, 0, jst)

	if year == 0 {
		latest, err := models.Events(
			qm.OrderBy("time DESC"),
			qm.Limit(1),
		).One(ctx, tx)
		if err == nil {
			start = latest.Time
		} else {
			start = time.Date(2015, time.January, 1, 0, 0, 0, 0, jst)
		}
		end = time.Now()
	}

	repo := repository.New(tx, fiat)

	start = start.Add(time.Second) // XXX
	for name, trans := range translators {
		log.Println("translate", name)
		err := trans.Translate(ctx, repo, start, end)
		if err != nil {
			return err
		}
	}
	return nil
}
