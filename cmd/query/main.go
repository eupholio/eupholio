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
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"

	"github.com/eupholio/eupholio/pkg/cmdutil"
	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/querycmd"
)

var jst = time.FixedZone("Asia/Tokyo", 9*60*60)

func main() {
	err := Execute()
	if err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "query",
	Short: "query is a tool for analyzing portfolio data",
	Long:  `query is a tool for analyzing portfolio data`,
}

func init() {
	rootCmd.AddCommand(
		SummarizeCmd(),
		BalanceCmd(),
		TransactionCmd(),
	)
}

// Execute runs root command
func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
	return nil
}

func BalanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "balance",
		Short: "show balance",
		RunE: func(cmd *cobra.Command, args []string) error {
			year, err := cmd.Flags().GetInt("year")
			if err != nil {
				return err
			}
			symbol, err := cmd.Flags().GetString("symbol")
			if err != nil {
				return err
			}
			source, err := cmd.Flags().GetString("source")
			if err != nil {
				return err
			}
			format, err := cmd.Flags().GetString("format")
			if err != nil {
				return err
			}

			w := os.Stdout
			ctx := context.Background()
			db, err := cmdutil.OpenDB()
			if err != nil {
				return err
			}
			return WithTx(ctx, db, func(tx *sql.Tx) error {
				return querycmd.QueryBalance(ctx, w, tx, year, currency.Symbol(symbol), source, querycmd.OutputFormat(format))
			})
		},
	}
	cmd.Flags().Int("year", 0, "year")
	cmd.Flags().String("symbol", "JPY", "base currency symbol")
	cmd.Flags().String("source", "yahoofinance", "data source")
	cmd.Flags().String("format", "table", "output format")
	return cmd
}

func TransactionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "transaction",
		Short: "show transaction histroy",
		RunE: func(cmd *cobra.Command, args []string) error {
			year, err := cmd.Flags().GetInt("year")
			if err != nil {
				return err
			}
			baseCurrency, err := cmd.Flags().GetString("symbol")
			if err != nil {
				return err
			}
			source, err := cmd.Flags().GetString("source")
			if err != nil {
				return err
			}
			format, err := cmd.Flags().GetString("format")
			if err != nil {
				return err
			}

			w := os.Stdout
			ctx := context.Background()
			db, err := cmdutil.OpenDB()
			if err != nil {
				return err
			}
			return WithTx(ctx, db, func(tx *sql.Tx) error {
				return querycmd.QueryTransactions(ctx, w, tx, year, jst, baseCurrency, source, querycmd.OutputFormat(format))
			})
		},
	}
	cmd.Flags().Int("year", 0, "year")
	cmd.Flags().String("symbol", "JPY", "base currency symbol")
	cmd.Flags().String("source", "yahoofinance", "data source")
	cmd.Flags().String("format", "table", "output format")
	return cmd
}

// WithTx runs fn with a transaction
func WithTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	return cmdutil.WithTx(ctx, db, fn)
}
