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
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/cmdutil"
)

func SummarizeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "summarize",
		Short: "summarieze data",
	}
	cmd.AddCommand(summarizeCurrencyCmd())
	return cmd
}

func summarizeCurrencyCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "currency",
		Short: "summarize currency",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := os.Stdout
			ctx := context.Background()

			db, err := cmdutil.OpenDB()
			if err != nil {
				return err
			}

			events, err := models.Events(qm.Select("currency"), qm.GroupBy("currency")).All(ctx, db)
			if err != nil {
				return err
			}

			for _, event := range events {
				fmt.Fprintf(w, "%s\n", event.Currency)
			}
			return nil
		},
	}
	return cmd
}
