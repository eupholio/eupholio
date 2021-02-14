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

package mam

import (
	"github.com/ericlagergren/decimal"
)

type aggregationContext struct {
	price     *decimal.Big
	beginning *decimal.Big
	quantity  *decimal.Big
	profit    *decimal.Big

	closeAmount   *decimal.Big
	closeQuantity *decimal.Big
	openAmount    *decimal.Big
	openQuantity  *decimal.Big
}

func NewAggregationContext(price *decimal.Big, quantity *decimal.Big) *aggregationContext {
	return &aggregationContext{
		price:         new(decimal.Big).Copy(price),
		beginning:     new(decimal.Big).Copy(quantity),
		quantity:      new(decimal.Big).Copy(quantity),
		profit:        decimal.New(0, 0),
		closeAmount:   decimal.New(0, 0),
		closeQuantity: decimal.New(0, 0),
		openAmount:    decimal.New(0, 0),
		openQuantity:  decimal.New(0, 0),
	}
}

func (c *aggregationContext) Price() *decimal.Big {
	if c.price.IsInf(0) {
		return decimal.New(0, 0)
	}
	return new(decimal.Big).Copy(c.price)
}

func (c *aggregationContext) Beginning() *decimal.Big {
	return new(decimal.Big).Copy(c.beginning)
}

func (c *aggregationContext) Quantity() *decimal.Big {
	return new(decimal.Big).Copy(c.quantity)
}

func (c *aggregationContext) Profit() *decimal.Big {
	return new(decimal.Big).Copy(c.profit)
}

func (c *aggregationContext) CloseQuantity() *decimal.Big {
	return new(decimal.Big).Copy(c.closeQuantity)
}

func (c *aggregationContext) OpenQuantity() *decimal.Big {
	return new(decimal.Big).Copy(c.openQuantity)
}

func (c *aggregationContext) ProcessOpen(quantity, fiatAmount *decimal.Big) {
	// average price
	q := new(decimal.Big).Add(c.quantity, quantity) // q = quantity + buy quantity
	a := new(decimal.Big).Mul(c.quantity, c.price)  // a = quantity * price
	t := new(decimal.Big).Add(a, fiatAmount)        // t = (quantity * price) + fiat amount
	c.price.Quo(t, q)                               // price = ((quantity * price) + fiat amount) / (quantity + buy quantity)
	c.quantity.Add(c.quantity, quantity)            // quantity = quantity + buy quantity
	// buy amount and quantity
	c.openAmount.Add(c.openAmount, fiatAmount)
	c.openQuantity.Add(c.openQuantity, quantity)
}

func (c *aggregationContext) ProcessClose(quantity, fiatAmount *decimal.Big) {
	// realized profit
	c.quantity.Sub(c.quantity, quantity)         // quantity = quantity - sell quantity
	p := new(decimal.Big).Mul(c.price, quantity) // p = price * sell quantity
	c.profit.Add(c.profit, p.Sub(fiatAmount, p)) // profit[currency] += fiat amount - (price * sell quantity)
	// sell amount and quantity
	c.closeAmount.Add(c.closeAmount, fiatAmount)
	c.closeQuantity.Add(c.closeQuantity, quantity)
}
