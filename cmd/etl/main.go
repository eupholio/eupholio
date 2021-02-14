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
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/spf13/cobra"
	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/eupholio/eupholio/pkg/cmdutil"
)

func main() {
	if debug, ok := os.LookupEnv("DEBUG"); ok {
		if strings.ToLower(debug) == "true" {
			boil.DebugMode = true
		}
	}
	err := Execute()
	if err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "etl",
	Short: "etl is a ETL tool to process raw transaction data",
	Long:  `etl is a ETL tool to process raw transaction data`,
}

func init() {
	rootCmd.AddCommand(
		LoadCmd(),
		ImportCmd(),
		CalculateCmd(),
		TranslateCmd(),
		DownloadCmd(),
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

// OpenDB opens a database with default settings
func OpenDB() (*sql.DB, error) {
	return cmdutil.OpenDB()
}

// WithTx runs fn with a transaction
func WithTx(ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) error) error {
	return cmdutil.WithTx(ctx, db, fn)
}
