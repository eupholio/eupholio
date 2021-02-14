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

package wam

import (
	"fmt"

	"github.com/ericlagergren/decimal"

	"github.com/eupholio/eupholio/models"
	"github.com/eupholio/eupholio/pkg/eupholio"
)

type calculationContext struct {
	sellAmount       *decimal.Big
	sellQuantity     *decimal.Big
	buyAmount        *decimal.Big
	buyQuantity      *decimal.Big
	depositQuantity  *decimal.Big
	withdrawQuantity *decimal.Big
}

func NewCalculationContext() *calculationContext {
	return &calculationContext{
		sellAmount:       decimal.New(0, 0),
		sellQuantity:     decimal.New(0, 0),
		buyAmount:        decimal.New(0, 0),
		buyQuantity:      decimal.New(0, 0),
		depositQuantity:  decimal.New(0, 0),
		withdrawQuantity: decimal.New(0, 0),
	}
}

func (c *calculationContext) AppendEntry(entry *models.Entry) error {
	quantity := entry.Quantity.Big
	fiatAmount := entry.FiatQuantity.Big
	switch entry.Type {
	case eupholio.EntryTypeOpen:
		c.buyAmount.Add(c.buyAmount, fiatAmount)
		c.buyQuantity.Add(c.buyQuantity, quantity)
	case eupholio.EntryTypeClose:
		c.sellAmount.Add(c.sellAmount, fiatAmount)
		c.sellQuantity.Add(c.sellQuantity, quantity)
	default:
		return fmt.Errorf("unknown entry type: %s", entry.Type)
	}
	return nil
}
