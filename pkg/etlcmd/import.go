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
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/eupholio/eupholio/pkg/bitflyer"
	"github.com/eupholio/eupholio/pkg/bittrex"
	"github.com/eupholio/eupholio/pkg/coincheck"
	"github.com/eupholio/eupholio/pkg/cryptact"
	"github.com/eupholio/eupholio/pkg/eupholio"
	"github.com/eupholio/eupholio/pkg/poloniex"
)

func ImportBitflyerData(ctx context.Context, db boil.ContextExecutor, args []string, overwrite bool) error {
	executor := bitflyer.NewExecutor()

	var opts []eupholio.Option
	if overwrite {
		opts = append(opts, eupholio.OverwriteOption())
	}

	for _, arg := range args {
		err := extract(ctx, arg, db, executor, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func ImportCoincheckData(ctx context.Context, db boil.ContextExecutor, args []string, overwrite bool) error {
	executor := coincheck.NewExecutor()

	var opts []eupholio.Option
	if overwrite {
		opts = append(opts, eupholio.OverwriteOption())
	}

	for _, arg := range args {
		err := extract(ctx, arg, db, executor, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func ImportBittrexData(ctx context.Context, db boil.ContextExecutor, args []string, overwrite bool, filetype string) error {
	var opts []eupholio.Option
	if overwrite {
		opts = append(opts, eupholio.OverwriteOption())
	}

	var executor eupholio.Extractor
	switch filetype {
	case "order":
		executor = bittrex.NewExtractor()
	case "deposit":
		executor = bittrex.NewDepositExtractor()
	case "withdraw":
		executor = bittrex.NewWithdrawExtractor()
	default:
		return fmt.Errorf("unknown file type: %s", filetype)
	}

	for _, arg := range args {
		err := extract(ctx, arg, db, executor, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func ImportPoloniexData(ctx context.Context, db boil.ContextExecutor, args []string, overwrite bool, filetype string) error {
	var opts []eupholio.Option
	if overwrite {
		opts = append(opts, eupholio.OverwriteOption())
	}

	extractors := map[string]eupholio.Extractor{
		"trades":        poloniex.NewTradeExtractor(),
		"deposits":      poloniex.NewDepositExtractor(),
		"withdrawals":   poloniex.NewWithdrawalExtractor(),
		"distributions": poloniex.NewDistributionExtractor(),
	}

	for _, arg := range args {
		var executor eupholio.Extractor
		ft := filetype
		if len(ft) == 0 {
			ss := strings.Split(filepath.Base(arg), "-")
			if len(ss) > 0 {
				if _, ok := extractors[ss[0]]; ok {
					ft = ss[0]
				} else {
					continue
				}
			}
		}
		executor, ok := extractors[ft]
		if !ok {
			return fmt.Errorf("unknown file type: %s", ft)
		}
		err := extract(ctx, arg, db, executor, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func ImportCryptactData(ctx context.Context, db boil.ContextExecutor, args []string, overwrite bool, filetype string, location string) error {
	var opts []eupholio.Option
	if overwrite {
		opts = append(opts, eupholio.OverwriteOption())
	}

	loc, err := time.LoadLocation(location)
	if err != nil {
		return err
	}

	var executor eupholio.Extractor
	switch filetype {
	case "custom":
		executor = cryptact.NewExtractor(loc)
	default:
		return fmt.Errorf("unknown file type: %s", filetype)
	}

	for _, arg := range args {
		err := extract(ctx, arg, db, executor, opts)
		if err != nil {
			return err
		}
	}
	return nil
}

func extract(ctx context.Context, path string, db boil.ContextExecutor, extractor eupholio.Extractor, opts []eupholio.Option) error {
	reader, err := os.Open(path)
	if err != nil {
		return err
	}
	defer reader.Close()
	err = extractor.Execute(ctx, db, reader, opts...)
	if err != nil {
		return err
	}
	return nil
}
