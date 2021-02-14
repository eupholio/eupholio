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

package cryptact

const WalletCode = "CRYPTACT_C"

// Actions (BUY/SELL/PAY/MINING/SENDFEE/TIP/REDUCE/BONUS/LENDING/STAKING)
const (
	ActionBuy     = "BUY"
	ActionSell    = "SELL"
	ActionPay     = "PAY"
	ActionMining  = "MINING"
	ActionSendFee = "SENDFEE"
	ActionTip     = "TIP"
	ActionReduce  = "REDUCE"
	ActionBonus   = "BONUS"
	ActionLending = "LENDING"
	ActionStaking = "STAKING"
)

// Column names
const (
	TimestampColumn = "Timestamp"
	ActionColumn    = "Action"
	SourceColumn    = "Source"
	BaseColumn      = "Base"
	VolumeColumn    = "Volume"
	PriceColumn     = "Price"
	CounterColumn   = "Counter"
	FeeColumn       = "Fee"
	FeeCcyColumn    = "FeeCcy"
	CommentColumn   = "Comment"
)

var columnNames = []string{
	TimestampColumn,
	ActionColumn,
	SourceColumn,
	BaseColumn,
	VolumeColumn,
	PriceColumn,
	CounterColumn,
	FeeColumn,
	FeeCcyColumn,
	CommentColumn,
}

var columnNamesSet map[string]struct{}

func init() {
	columnNamesSet = make(map[string]struct{})
	for _, name := range columnNames {
		columnNamesSet[name] = struct{}{}
	}
}
