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

package httputil

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

var ErrNotFound = errors.New("not found")

func HttpGet(ctx context.Context, url string, timeout time.Duration) (content []byte, err error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return
	}

	ctx, cancel_func := context.WithTimeout(ctx, timeout)
	defer cancel_func()
	request = request.WithContext(ctx)

	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		if response.StatusCode == 404 {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("INVALID RESPONSE; status: %s", response.Status)
	}

	return ioutil.ReadAll(response.Body)
}
