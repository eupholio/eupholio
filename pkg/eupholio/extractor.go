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
	"io"

	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Config struct {
	Overwrite bool
	Debug     bool
}

type Option func(config *Config)

func OverwriteOption() Option {
	return func(config *Config) {
		config.Overwrite = true
	}
}

func DebugOption() Option {
	return func(config *Config) {
		config.Debug = true
	}
}

type Extractor interface {
	Execute(ctx context.Context, db boil.ContextExecutor, reader io.Reader, options ...Option) error
}
