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

package eupholio

import (
	"context"
	"sort"
	"time"

	"github.com/volatiletech/sqlboiler/v4/boil"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/currency"
)

type ConfigRepository interface {
	FindConfigByYear(ctx context.Context, year int) (*models.Config, error)
}

type TransactionRepository interface {
	CreateTransaction(ctx context.Context, time time.Time, exchangeCode string, id int) (*models.Transaction, error)
	DeleteTransaction(ctx context.Context, exchangeCode string, start, end time.Time) (int64, error)
	FindTransactionsByYear(ctx context.Context, year int, location *time.Location) (models.TransactionSlice, error)
}

type EventRepository interface {
	CreateEvents(ctx context.Context, events models.EventSlice) error
	FindEvents(ctx context.Context) (models.EventSlice, error)
	FindEventsByYear(ctx context.Context, year int, location *time.Location) (models.EventSlice, error)
	FindEventsByStartAndEnd(ctx context.Context, start, end time.Time) (models.EventSlice, error)
	FindEventsByCurrencyStartAndEnd(ctx context.Context, currency currency.Symbol, start, end time.Time) (models.EventSlice, error)
}

type EntryRepository interface {
	CreateEntries(ctx context.Context, entries models.EntrySlice) error
	UpdateEntries(ctx context.Context, entries models.EntrySlice) error
	FindEntriesByYear(ctx context.Context, year int, loc *time.Location) (models.EntrySlice, error)
	FindEntriesByStartAndEnd(ctx context.Context, start, end time.Time) (models.EntrySlice, error)
}

type MarketPriceRepository interface {
	CreateMarketPrices(ctx context.Context, marketPrices models.MarketPriceSlice) error
	AppendMarketPrices(ctx context.Context, marketPrices models.MarketPriceSlice) error
	FindLatestMarketPriceByCurrency(ctx context.Context, currency string) (*models.MarketPrice, error)
	FindMarketPriceByCurrencyAndTime(ctx context.Context, currency string, time time.Time) (*models.MarketPrice, error)
}

type BalanceRepository interface {
	CreateBalances(ctx context.Context, balances models.BalanceSlice) error
	FindBalanceByCurrencyAndYear(ctx context.Context, currency string, year int) (*models.Balance, error)
	FindBalancesByYear(ctx context.Context, year int) (models.BalanceSlice, error)
}

type Repository interface {
	boil.ContextExecutor
	ConfigRepository
	TransactionRepository
	EventRepository
	EntryRepository
	MarketPriceRepository
	BalanceRepository
}

type EventsOfTransaction struct {
	ID          int
	Time        time.Time
	WalletCode  string
	Events      models.EventSlice
	Description string
}

func FindEventsOfTransactions(ctx context.Context, repo interface {
	TransactionRepository
	EventRepository
}, year int, loc *time.Location) ([]*EventsOfTransaction, error) {
	transactions, err := repo.FindTransactionsByYear(ctx, year, loc)
	if err != nil {
		return nil, err
	}
	trMap := make(map[int]*EventsOfTransaction)
	for _, t := range transactions {
		trMap[t.ID] = &EventsOfTransaction{
			ID:          t.ID,
			Time:        t.Time,
			WalletCode:  t.WalletCode,
			Description: t.Description,
		}
	}
	events, err := repo.FindEventsByYear(ctx, year, loc)
	if err != nil {
		return nil, err
	}
	for _, e := range events {
		t, ok := trMap[e.TransactionID]
		if ok {
			t.Events = append(t.Events, e)
		}
	}

	var ret []*EventsOfTransaction
	for _, v := range trMap {
		ret = append(ret, v)
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].Time.Before(ret[j].Time)
	})
	return ret, nil
}

type EntriesOfTransaction struct {
	ID          int
	Time        time.Time
	WalletCode  string
	Entries     models.EntrySlice
	Description string
}

func FindEntriesOfTransactions(ctx context.Context, repo interface {
	TransactionRepository
	EntryRepository
}, year int, loc *time.Location) ([]*EntriesOfTransaction, error) {
	transactions, err := repo.FindTransactionsByYear(ctx, year, loc)
	if err != nil {
		return nil, err
	}
	trMap := make(map[int]*EntriesOfTransaction)
	for _, t := range transactions {
		trMap[t.ID] = &EntriesOfTransaction{
			ID:          t.ID,
			Time:        t.Time,
			WalletCode:  t.WalletCode,
			Description: t.Description,
		}
	}
	entries, err := repo.FindEntriesByYear(ctx, year, loc)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		t, ok := trMap[e.TransactionID]
		if ok {
			t.Entries = append(t.Entries, e)
		}
	}

	var ret []*EntriesOfTransaction
	for _, v := range trMap {
		ret = append(ret, v)
	}
	sort.SliceStable(ret, func(i, j int) bool {
		return ret[i].Time.Before(ret[j].Time)
	})
	return ret, nil
}
