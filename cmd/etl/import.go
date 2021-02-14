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

// ImportCmd imports data from files
func ImportCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import",
		Short: "import data",
	}
	cmd.AddCommand(
		importBitflyerCmd(),
		importCoincheckCmd(),
		importBittrexCmd(),
		importPoloniexCmd(),
		importCryptactCmd(),
	)
	return cmd
}

func importBitflyerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bf",
		Short: "import bitFlyer data",
		RunE: func(cmd *cobra.Command, args []string) error {
			overwrite, err := cmd.Flags().GetBool("overwrite")
			if err != nil {
				return err
			}
			db, err := OpenDB()
			if err != nil {
				return err
			}
			ctx := context.Background()
			return WithTx(ctx, db, func(tx *sql.Tx) error {
				return etlcmd.ImportBitflyerData(ctx, db, args, overwrite)
			})
		},
	}
	cmd.Flags().Bool("overwrite", false, "overwrite")
	return cmd
}

func importCoincheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coincheck",
		Short: "import coincheck data",
		RunE: func(cmd *cobra.Command, args []string) error {
			overwrite, err := cmd.Flags().GetBool("overwrite")
			if err != nil {
				return err
			}
			db, err := OpenDB()
			if err != nil {
				return err
			}
			ctx := context.Background()
			return WithTx(ctx, db, func(tx *sql.Tx) error {
				return etlcmd.ImportCoincheckData(ctx, db, args, overwrite)
			})
		},
	}
	cmd.Flags().Bool("overwrite", false, "overwrite")
	return cmd
}

func importBittrexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bittrex",
		Short: "import bittrex data",
		RunE: func(cmd *cobra.Command, args []string) error {
			overwrite, err := cmd.Flags().GetBool("overwrite")
			if err != nil {
				return err
			}
			filetype, err := cmd.Flags().GetString("filetype")
			if err != nil {
				return err
			}
			db, err := OpenDB()
			if err != nil {
				return err
			}
			ctx := context.Background()
			return WithTx(ctx, db, func(tx *sql.Tx) error {
				return etlcmd.ImportBittrexData(ctx, db, args, overwrite, filetype)
			})
		},
	}
	cmd.Flags().Bool("overwrite", false, "overwrite")
	cmd.Flags().String("filetype", "order", "file type (order, deposit, withdraw)")
	return cmd
}

func importPoloniexCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "poloniex",
		Short: "import poloniex data",
		RunE: func(cmd *cobra.Command, args []string) error {
			overwrite, err := cmd.Flags().GetBool("overwrite")
			if err != nil {
				return err
			}
			filetype, err := cmd.Flags().GetString("filetype")
			if err != nil {
				return err
			}
			db, err := OpenDB()
			if err != nil {
				return err
			}
			ctx := context.Background()
			return WithTx(ctx, db, func(tx *sql.Tx) error {
				return etlcmd.ImportPoloniexData(ctx, db, args, overwrite, filetype)
			})
		},
	}
	cmd.Flags().Bool("overwrite", false, "overwrite")
	cmd.Flags().String("filetype", "", "file type (trades, deposits, withdrawals, distributions)")
	return cmd
}

func importCryptactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cryptact",
		Short: "import cryptact data",
		RunE: func(cmd *cobra.Command, args []string) error {
			overwrite, err := cmd.Flags().GetBool("overwrite")
			if err != nil {
				return err
			}
			filetype, err := cmd.Flags().GetString("filetype")
			if err != nil {
				return err
			}
			timezone, err := cmd.Flags().GetString("timezone")
			if err != nil {
				return err
			}
			db, err := OpenDB()
			if err != nil {
				return err
			}
			ctx := context.Background()
			return WithTx(ctx, db, func(tx *sql.Tx) error {
				return etlcmd.ImportCryptactData(ctx, db, args, overwrite, filetype, timezone)
			})
		},
	}
	cmd.Flags().Bool("overwrite", false, "overwrite")
	cmd.Flags().String("filetype", "custom", "file type (custom)")
	cmd.Flags().String("timezone", "UTC", "timezone (UTC)")
	return cmd
}
