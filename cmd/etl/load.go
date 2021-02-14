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
	"context"
	"database/sql"

	"github.com/spf13/cobra"

	"github.com/eupholio/eupholio/pkg/etlcmd"
)

// LoadCmd load master data from files
func LoadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
		Short: "load master data",
	}
	cmd.AddCommand(
		loadCoingeckoCmd(),
		loadYahooFinanceCmd(),
		loadCDDCmd(),
	)
	return cmd
}

func loadCoingeckoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coingecko",
		Short: "load coingecko data",
	}
	cmd.AddCommand(
		loadCoingeckoHistoricalPriceCmd(),
	)
	return cmd
}

func loadCoingeckoHistoricalPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "historical_price",
		Short: "load Coingecko historical price data",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := OpenDB()
			if err != nil {
				return err
			}
			ctx := context.Background()
			return WithTx(ctx, db, func(tx *sql.Tx) error {
				return etlcmd.LoadCoingeckoHistoricalPrice(db, args)
			})
		},
	}
	cmd.Flags().Bool("overwrite", false, "overwrite")
	return cmd
}

func loadYahooFinanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yahoofinance",
		Short: "load Yahoo Finance data",
	}
	cmd.AddCommand(loadYahooFinanceHistoricalPriceCmd())
	return cmd
}

func loadYahooFinanceHistoricalPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "historical_price",
		Short: "load Yahoo Finance historical price data",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := OpenDB()
			if err != nil {
				return err
			}
			ctx := context.Background()
			return WithTx(ctx, db, func(tx *sql.Tx) error {
				return etlcmd.LoadYahooFinanceHistoricalPrice(db, args)
			})
		},
	}
	cmd.Flags().Bool("overwrite", false, "overwrite")
	return cmd
}

func loadCDDCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cryptodatadownload",
		Short: "load CryptoDataDownload data",
	}
	cmd.AddCommand(loadCDDHistoricalPriceCmd())
	return cmd
}

func loadCDDHistoricalPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "historical_price",
		Short: "load CryptoDataDownload historical price data",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := OpenDB()
			if err != nil {
				return err
			}
			ctx := context.Background()
			for _, arg := range args {
				err := WithTx(ctx, db, func(tx *sql.Tx) error {
					return etlcmd.LoadCDDHistoricalPrice(db, []string{arg})
				})
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.Flags().Bool("overwrite", false, "overwrite")
	return cmd
}
