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

package poloniex

import (
	"github.com/eupholio/eupholio/pkg/eupholio"
	"github.com/eupholio/eupholio/pkg/poloniex/deposit"
	"github.com/eupholio/eupholio/pkg/poloniex/distribution"
	"github.com/eupholio/eupholio/pkg/poloniex/repository"
	"github.com/eupholio/eupholio/pkg/poloniex/trade"
	"github.com/eupholio/eupholio/pkg/poloniex/withdrawal"
)

const WalletCode = "POLONIEX"

type Repository = repository.Repository

func NewTradeExtractor() eupholio.Extractor {
	return trade.NewExtractor()
}

func NewDepositExtractor() eupholio.Extractor {
	return deposit.NewExtractor()
}

func NewWithdrawalExtractor() eupholio.Extractor {
	return withdrawal.NewExtractor()
}

func NewDistributionExtractor() eupholio.Extractor {
	return distribution.NewExtractor()
}
