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

	"github.com/eupholio/eupholio/models"
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
	Use:   "config",
	Short: "config is a configuration tool",
	Long:  `config is a configuration tool`,
}

func init() {
	rootCmd.AddCommand(configCostMethodCmd())
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

func configCostMethodCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "costmethod",
		Short: "set cost calcuration method",
		RunE: func(cmd *cobra.Command, args []string) error {
			db, err := OpenDB()
			if err != nil {
				return err
			}
			year, err := cmd.Flags().GetInt("year")
			if err != nil {
				return err
			}
			method, err := cmd.Flags().GetString("method")
			if err != nil {
				return err
			}
			ctx := context.Background()
			return WithTx(ctx, db, func(tx *sql.Tx) error {
				c, err := models.FindConfig(ctx, tx, 0, year)
				if err == sql.ErrNoRows {
					c = &models.Config{
						ID:         0,
						Year:       year,
						CostMethod: method,
					}
					return c.Insert(ctx, tx, boil.Infer())
				}
				c.CostMethod = method
				_, err = c.Update(ctx, tx, boil.Infer())
				if err != nil {
					return err
				}
				return nil
			})
		},
	}
	cmd.Flags().Int("year", 0, "year")
	cmd.Flags().String("method", "", "method")
	return cmd
}
