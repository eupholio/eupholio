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
	"time"

	"github.com/spf13/cobra"

	"github.com/eupholio/eupholio/pkg/costmethod"
	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/etlcmd"
)

// CalculateCmd imports data from files
func CalculateCmd() *cobra.Command {
	jst := time.FixedZone("Asia/Tokyo", 9*60*60)
	cmd := &cobra.Command{
		Use:   "calculate",
		Short: "calculate profit",
		RunE: func(cmd *cobra.Command, args []string) error {
			year, err := cmd.Flags().GetInt("year")
			if err != nil {
				return err
			}
			fiat, err := cmd.Flags().GetString("fiat")
			if err != nil {
				return err
			}
			method, err := cmd.Flags().GetString("method")
			if err != nil {
				return err
			}

			var options []costmethod.Option
			debug, err := cmd.Flags().GetBool("debug")
			if debug {
				options = append(options, costmethod.DebugOption())
			}

			db, err := OpenDB()
			if err != nil {
				return err
			}

			ctx := context.Background()
			return WithTx(ctx, db, func(tx *sql.Tx) error {
				return etlcmd.Calculate(ctx, tx, year, currency.Symbol(fiat), jst, method, options...)
			})
		},
	}
	cmd.Flags().Bool("debug", false, "debug")
	cmd.Flags().Int("year", 0, "year")
	cmd.Flags().String("fiat", "JPY", "fiat currency ticker code")
	cmd.Flags().String("method", "", "override cost calculation method (wam, mam)")
	return cmd
}
