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
	"github.com/spf13/cobra"
	
	"github.com/eupholio/eupholio/pkg/currency"
	"github.com/eupholio/eupholio/pkg/etlcmd"
)

func DownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "download",
		Short: "download master data",
	}
	cmd.AddCommand(
		downloadCoingeckoCmd(),
		downloadCryptoDataDownloadCmd(),
		downloadYahooFinanceCmd(),
	)
	return cmd
}

func downloadCoingeckoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "coingecko",
		Short: "download coingecko data",
	}
	cmd.AddCommand(downloadCoingeckoHistoricalPriceCmd())
	return cmd
}

func downloadCoingeckoHistoricalPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "historical_price",
		Short: "download Coingecko historical price data",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cmd.Flags().GetString("dir")
			if err != nil {
				return err
			}
			return etlcmd.DownloadCoingeckoHistoricalPrice(dir)
		},
	}
	cmd.Flags().String("dir", "pricedata/coingecko", "output directory")
	return cmd
}

func downloadCryptoDataDownloadCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cryptodatadownload",
		Short: "download Crypto Data Download data",
	}
	cmd.AddCommand(downloadCryptoDataDownloadHistoricalPriceCmd())
	return cmd
}

func downloadCryptoDataDownloadHistoricalPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "historical_price",
		Short: "download Crypto Data Download historical price data",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cmd.Flags().GetString("dir")
			if err != nil {
				return err
			}
			return etlcmd.DownloadCryptoDataDownloadHistoricalPrice(dir)
		},
	}
	cmd.Flags().String("dir", "pricedata/cryptodatadownload", "output directory")
	return cmd
}

func downloadYahooFinanceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yahoofinance",
		Short: "download Yahoo Finance data",
	}
	cmd.AddCommand(downloadYahooFinanceHistoricalPriceCmd())
	return cmd
}

func downloadYahooFinanceHistoricalPriceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "historical_price",
		Short: "download historical price data from Yahoo Finance",
		RunE: func(cmd *cobra.Command, args []string) error {
			dir, err := cmd.Flags().GetString("dir")
			if err != nil {
				return err
			}
			fiat, err := cmd.Flags().GetStringSlice("fiat")
			if err != nil {
				return err
			}
			symbols, err := cmd.Flags().GetStringSlice("symbol")
			if err != nil {
				return err
			}
			return etlcmd.DownloadYahooFinanceHistoricalPrice(dir, fiat, symbols)
		},
	}
	cmd.Flags().String("dir", "pricedata/yahoofinance", "output directory")
	cmd.Flags().StringSlice("fiat", currency.FiatCurrencies.Strings(), "fiat currency")
	cmd.Flags().StringSlice("symbol", currency.BaseCurrencies.Strings(), "symbols")
	return cmd
}
