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

package costmethod

import (
	"fmt"
	"strings"

	"github.com/ericlagergren/decimal"
)

type balance struct {
	Position *decimal.Big
	Open     *decimal.Big
	Close    *decimal.Big
	Init     *decimal.Big
}

type CalculateContext struct {
	balances map[string]*balance
}

func NewCaluculateContext() *CalculateContext {
	return &CalculateContext{
		balances: make(map[string]*balance),
	}
}

func (c *CalculateContext) InitPosition(currency string, quantity *decimal.Big) {
	c.updatePosition(currency, quantity, nil)
	b := c.balances[currency]
	b.Init.Copy(quantity)
}

func (c *CalculateContext) OpenPosition(currency string, quantity *decimal.Big) {
	c.updatePosition(currency, quantity, nil)
	b := c.balances[currency]
	b.Open.Add(b.Open, quantity)
}

func (c *CalculateContext) updatePosition(currency string, open, close *decimal.Big) {
	b, ok := c.balances[currency]
	if !ok {
		b = &balance{
			Position: decimal.New(0, 0),
			Open:     decimal.New(0, 0),
			Close:    decimal.New(0, 0),
			Init:     decimal.New(0, 0),
		}
		c.balances[currency] = b
	}
	if open != nil {
		b.Position.Add(b.Position, open)
	}
	if close != nil {
		b.Position.Sub(b.Position, close)
	}
}

func (c *CalculateContext) ClosePosition(currency string, quantity *decimal.Big) {
	c.updatePosition(currency, nil, quantity)
	b := c.balances[currency]
	b.Close.Add(b.Close, quantity)
}

func (c *CalculateContext) Position(currency string) *decimal.Big {
	b, ok := c.balances[currency]
	if !ok {
		return decimal.New(0, 0)
	}
	return new(decimal.Big).Copy(b.Position)
}

func (c *CalculateContext) String() string {
	var sb strings.Builder
	first := true
	for c, b := range c.balances {
		if !first {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("%s %v = %v + %v - %v", c, b.Position, b.Init, b.Open, b.Close))
		first = false
	}
	return sb.String()
}

func (c *CalculateContext) Balances() map[string]*balance {
	return c.balances
}

func (c *CalculateContext) Balance(currency string) (*balance, bool) {
	b, ok := c.balances[currency]
	return b, ok
}
