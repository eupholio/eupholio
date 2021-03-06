// Code generated by SQLBoiler 4.4.0 (https://github.com/volatiletech/sqlboiler). DO NOT EDIT.
// This file is meant to be re-generated in place and/or deleted at any time.

package models

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/strmangle"
)

// CoincheckHistory is an object representing the database table.
type CoincheckHistory struct {
	ID               int               `boil:"id" json:"id" toml:"id" yaml:"id"`
	IDCode           string            `boil:"id_code" json:"id_code" toml:"id_code" yaml:"id_code"`
	Time             time.Time         `boil:"time" json:"time" toml:"time" yaml:"time"`
	Operation        string            `boil:"operation" json:"operation" toml:"operation" yaml:"operation"`
	Amount           types.Decimal     `boil:"amount" json:"amount" toml:"amount" yaml:"amount"`
	TradingCurrency  string            `boil:"trading_currency" json:"trading_currency" toml:"trading_currency" yaml:"trading_currency"`
	Price            types.NullDecimal `boil:"price" json:"price,omitempty" toml:"price" yaml:"price,omitempty"`
	OriginalCurrency null.String       `boil:"original_currency" json:"original_currency,omitempty" toml:"original_currency" yaml:"original_currency,omitempty"`
	Fee              types.NullDecimal `boil:"fee" json:"fee,omitempty" toml:"fee" yaml:"fee,omitempty"`
	Comment          string            `boil:"comment" json:"comment" toml:"comment" yaml:"comment"`

	R *coincheckHistoryR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L coincheckHistoryL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var CoincheckHistoryColumns = struct {
	ID               string
	IDCode           string
	Time             string
	Operation        string
	Amount           string
	TradingCurrency  string
	Price            string
	OriginalCurrency string
	Fee              string
	Comment          string
}{
	ID:               "id",
	IDCode:           "id_code",
	Time:             "time",
	Operation:        "operation",
	Amount:           "amount",
	TradingCurrency:  "trading_currency",
	Price:            "price",
	OriginalCurrency: "original_currency",
	Fee:              "fee",
	Comment:          "comment",
}

// Generated where

var CoincheckHistoryWhere = struct {
	ID               whereHelperint
	IDCode           whereHelperstring
	Time             whereHelpertime_Time
	Operation        whereHelperstring
	Amount           whereHelpertypes_Decimal
	TradingCurrency  whereHelperstring
	Price            whereHelpertypes_NullDecimal
	OriginalCurrency whereHelpernull_String
	Fee              whereHelpertypes_NullDecimal
	Comment          whereHelperstring
}{
	ID:               whereHelperint{field: "`coincheck_history`.`id`"},
	IDCode:           whereHelperstring{field: "`coincheck_history`.`id_code`"},
	Time:             whereHelpertime_Time{field: "`coincheck_history`.`time`"},
	Operation:        whereHelperstring{field: "`coincheck_history`.`operation`"},
	Amount:           whereHelpertypes_Decimal{field: "`coincheck_history`.`amount`"},
	TradingCurrency:  whereHelperstring{field: "`coincheck_history`.`trading_currency`"},
	Price:            whereHelpertypes_NullDecimal{field: "`coincheck_history`.`price`"},
	OriginalCurrency: whereHelpernull_String{field: "`coincheck_history`.`original_currency`"},
	Fee:              whereHelpertypes_NullDecimal{field: "`coincheck_history`.`fee`"},
	Comment:          whereHelperstring{field: "`coincheck_history`.`comment`"},
}

// CoincheckHistoryRels is where relationship names are stored.
var CoincheckHistoryRels = struct {
}{}

// coincheckHistoryR is where relationships are stored.
type coincheckHistoryR struct {
}

// NewStruct creates a new relationship struct
func (*coincheckHistoryR) NewStruct() *coincheckHistoryR {
	return &coincheckHistoryR{}
}

// coincheckHistoryL is where Load methods for each relationship are stored.
type coincheckHistoryL struct{}

var (
	coincheckHistoryAllColumns            = []string{"id", "id_code", "time", "operation", "amount", "trading_currency", "price", "original_currency", "fee", "comment"}
	coincheckHistoryColumnsWithoutDefault = []string{"id_code", "time", "operation", "amount", "trading_currency", "price", "original_currency", "fee", "comment"}
	coincheckHistoryColumnsWithDefault    = []string{"id"}
	coincheckHistoryPrimaryKeyColumns     = []string{"id"}
)

type (
	// CoincheckHistorySlice is an alias for a slice of pointers to CoincheckHistory.
	// This should generally be used opposed to []CoincheckHistory.
	CoincheckHistorySlice []*CoincheckHistory
	// CoincheckHistoryHook is the signature for custom CoincheckHistory hook methods
	CoincheckHistoryHook func(context.Context, boil.ContextExecutor, *CoincheckHistory) error

	coincheckHistoryQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	coincheckHistoryType                 = reflect.TypeOf(&CoincheckHistory{})
	coincheckHistoryMapping              = queries.MakeStructMapping(coincheckHistoryType)
	coincheckHistoryPrimaryKeyMapping, _ = queries.BindMapping(coincheckHistoryType, coincheckHistoryMapping, coincheckHistoryPrimaryKeyColumns)
	coincheckHistoryInsertCacheMut       sync.RWMutex
	coincheckHistoryInsertCache          = make(map[string]insertCache)
	coincheckHistoryUpdateCacheMut       sync.RWMutex
	coincheckHistoryUpdateCache          = make(map[string]updateCache)
	coincheckHistoryUpsertCacheMut       sync.RWMutex
	coincheckHistoryUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var coincheckHistoryBeforeInsertHooks []CoincheckHistoryHook
var coincheckHistoryBeforeUpdateHooks []CoincheckHistoryHook
var coincheckHistoryBeforeDeleteHooks []CoincheckHistoryHook
var coincheckHistoryBeforeUpsertHooks []CoincheckHistoryHook

var coincheckHistoryAfterInsertHooks []CoincheckHistoryHook
var coincheckHistoryAfterSelectHooks []CoincheckHistoryHook
var coincheckHistoryAfterUpdateHooks []CoincheckHistoryHook
var coincheckHistoryAfterDeleteHooks []CoincheckHistoryHook
var coincheckHistoryAfterUpsertHooks []CoincheckHistoryHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *CoincheckHistory) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range coincheckHistoryBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *CoincheckHistory) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range coincheckHistoryBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *CoincheckHistory) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range coincheckHistoryBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *CoincheckHistory) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range coincheckHistoryBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *CoincheckHistory) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range coincheckHistoryAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *CoincheckHistory) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range coincheckHistoryAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *CoincheckHistory) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range coincheckHistoryAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *CoincheckHistory) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range coincheckHistoryAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *CoincheckHistory) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range coincheckHistoryAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddCoincheckHistoryHook registers your hook function for all future operations.
func AddCoincheckHistoryHook(hookPoint boil.HookPoint, coincheckHistoryHook CoincheckHistoryHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		coincheckHistoryBeforeInsertHooks = append(coincheckHistoryBeforeInsertHooks, coincheckHistoryHook)
	case boil.BeforeUpdateHook:
		coincheckHistoryBeforeUpdateHooks = append(coincheckHistoryBeforeUpdateHooks, coincheckHistoryHook)
	case boil.BeforeDeleteHook:
		coincheckHistoryBeforeDeleteHooks = append(coincheckHistoryBeforeDeleteHooks, coincheckHistoryHook)
	case boil.BeforeUpsertHook:
		coincheckHistoryBeforeUpsertHooks = append(coincheckHistoryBeforeUpsertHooks, coincheckHistoryHook)
	case boil.AfterInsertHook:
		coincheckHistoryAfterInsertHooks = append(coincheckHistoryAfterInsertHooks, coincheckHistoryHook)
	case boil.AfterSelectHook:
		coincheckHistoryAfterSelectHooks = append(coincheckHistoryAfterSelectHooks, coincheckHistoryHook)
	case boil.AfterUpdateHook:
		coincheckHistoryAfterUpdateHooks = append(coincheckHistoryAfterUpdateHooks, coincheckHistoryHook)
	case boil.AfterDeleteHook:
		coincheckHistoryAfterDeleteHooks = append(coincheckHistoryAfterDeleteHooks, coincheckHistoryHook)
	case boil.AfterUpsertHook:
		coincheckHistoryAfterUpsertHooks = append(coincheckHistoryAfterUpsertHooks, coincheckHistoryHook)
	}
}

// One returns a single coincheckHistory record from the query.
func (q coincheckHistoryQuery) One(ctx context.Context, exec boil.ContextExecutor) (*CoincheckHistory, error) {
	o := &CoincheckHistory{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for coincheck_history")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all CoincheckHistory records from the query.
func (q coincheckHistoryQuery) All(ctx context.Context, exec boil.ContextExecutor) (CoincheckHistorySlice, error) {
	var o []*CoincheckHistory

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to CoincheckHistory slice")
	}

	if len(coincheckHistoryAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all CoincheckHistory records in the query.
func (q coincheckHistoryQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count coincheck_history rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q coincheckHistoryQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if coincheck_history exists")
	}

	return count > 0, nil
}

// CoincheckHistories retrieves all the records using an executor.
func CoincheckHistories(mods ...qm.QueryMod) coincheckHistoryQuery {
	mods = append(mods, qm.From("`coincheck_history`"))
	return coincheckHistoryQuery{NewQuery(mods...)}
}

// FindCoincheckHistory retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindCoincheckHistory(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*CoincheckHistory, error) {
	coincheckHistoryObj := &CoincheckHistory{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from `coincheck_history` where `id`=?", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, coincheckHistoryObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from coincheck_history")
	}

	return coincheckHistoryObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *CoincheckHistory) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no coincheck_history provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(coincheckHistoryColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	coincheckHistoryInsertCacheMut.RLock()
	cache, cached := coincheckHistoryInsertCache[key]
	coincheckHistoryInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			coincheckHistoryAllColumns,
			coincheckHistoryColumnsWithDefault,
			coincheckHistoryColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(coincheckHistoryType, coincheckHistoryMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(coincheckHistoryType, coincheckHistoryMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO `coincheck_history` (`%s`) %%sVALUES (%s)%%s", strings.Join(wl, "`,`"), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO `coincheck_history` () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT `%s` FROM `coincheck_history` WHERE %s", strings.Join(returnColumns, "`,`"), strmangle.WhereClause("`", "`", 0, coincheckHistoryPrimaryKeyColumns))
		}

		cache.query = fmt.Sprintf(cache.query, queryOutput, queryReturning)
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	result, err := exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "models: unable to insert into coincheck_history")
	}

	var lastID int64
	var identifierCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	lastID, err = result.LastInsertId()
	if err != nil {
		return ErrSyncFail
	}

	o.ID = int(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == coincheckHistoryMapping["id"] {
		goto CacheNoHooks
	}

	identifierCols = []interface{}{
		o.ID,
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, identifierCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, identifierCols...).Scan(queries.PtrsFromMapping(value, cache.retMapping)...)
	if err != nil {
		return errors.Wrap(err, "models: unable to populate default values for coincheck_history")
	}

CacheNoHooks:
	if !cached {
		coincheckHistoryInsertCacheMut.Lock()
		coincheckHistoryInsertCache[key] = cache
		coincheckHistoryInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the CoincheckHistory.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *CoincheckHistory) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	coincheckHistoryUpdateCacheMut.RLock()
	cache, cached := coincheckHistoryUpdateCache[key]
	coincheckHistoryUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			coincheckHistoryAllColumns,
			coincheckHistoryPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update coincheck_history, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE `coincheck_history` SET %s WHERE %s",
			strmangle.SetParamNames("`", "`", 0, wl),
			strmangle.WhereClause("`", "`", 0, coincheckHistoryPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(coincheckHistoryType, coincheckHistoryMapping, append(wl, coincheckHistoryPrimaryKeyColumns...))
		if err != nil {
			return 0, err
		}
	}

	values := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cache.valueMapping)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, values)
	}
	var result sql.Result
	result, err = exec.ExecContext(ctx, cache.query, values...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update coincheck_history row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for coincheck_history")
	}

	if !cached {
		coincheckHistoryUpdateCacheMut.Lock()
		coincheckHistoryUpdateCache[key] = cache
		coincheckHistoryUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q coincheckHistoryQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for coincheck_history")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for coincheck_history")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o CoincheckHistorySlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	ln := int64(len(o))
	if ln == 0 {
		return 0, nil
	}

	if len(cols) == 0 {
		return 0, errors.New("models: update all requires at least one column argument")
	}

	colNames := make([]string, len(cols))
	args := make([]interface{}, len(cols))

	i := 0
	for name, value := range cols {
		colNames[i] = name
		args[i] = value
		i++
	}

	// Append all of the primary key values for each column
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), coincheckHistoryPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE `coincheck_history` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, coincheckHistoryPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in coincheckHistory slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all coincheckHistory")
	}
	return rowsAff, nil
}

var mySQLCoincheckHistoryUniqueColumns = []string{
	"id",
	"id_code",
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *CoincheckHistory) Upsert(ctx context.Context, exec boil.ContextExecutor, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no coincheck_history provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(coincheckHistoryColumnsWithDefault, o)
	nzUniques := queries.NonZeroDefaultSet(mySQLCoincheckHistoryUniqueColumns, o)

	if len(nzUniques) == 0 {
		return errors.New("cannot upsert with a table that cannot conflict on a unique column")
	}

	// Build cache key in-line uglily - mysql vs psql problems
	buf := strmangle.GetBuffer()
	buf.WriteString(strconv.Itoa(updateColumns.Kind))
	for _, c := range updateColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	buf.WriteString(strconv.Itoa(insertColumns.Kind))
	for _, c := range insertColumns.Cols {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzDefaults {
		buf.WriteString(c)
	}
	buf.WriteByte('.')
	for _, c := range nzUniques {
		buf.WriteString(c)
	}
	key := buf.String()
	strmangle.PutBuffer(buf)

	coincheckHistoryUpsertCacheMut.RLock()
	cache, cached := coincheckHistoryUpsertCache[key]
	coincheckHistoryUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			coincheckHistoryAllColumns,
			coincheckHistoryColumnsWithDefault,
			coincheckHistoryColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			coincheckHistoryAllColumns,
			coincheckHistoryPrimaryKeyColumns,
		)

		if !updateColumns.IsNone() && len(update) == 0 {
			return errors.New("models: unable to upsert coincheck_history, could not build update column list")
		}

		ret = strmangle.SetComplement(ret, nzUniques)
		cache.query = buildUpsertQueryMySQL(dialect, "`coincheck_history`", update, insert)
		cache.retQuery = fmt.Sprintf(
			"SELECT %s FROM `coincheck_history` WHERE %s",
			strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, ret), ","),
			strmangle.WhereClause("`", "`", 0, nzUniques),
		)

		cache.valueMapping, err = queries.BindMapping(coincheckHistoryType, coincheckHistoryMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(coincheckHistoryType, coincheckHistoryMapping, ret)
			if err != nil {
				return err
			}
		}
	}

	value := reflect.Indirect(reflect.ValueOf(o))
	vals := queries.ValuesFromMapping(value, cache.valueMapping)
	var returns []interface{}
	if len(cache.retMapping) != 0 {
		returns = queries.PtrsFromMapping(value, cache.retMapping)
	}

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.query)
		fmt.Fprintln(writer, vals)
	}
	result, err := exec.ExecContext(ctx, cache.query, vals...)

	if err != nil {
		return errors.Wrap(err, "models: unable to upsert for coincheck_history")
	}

	var lastID int64
	var uniqueMap []uint64
	var nzUniqueCols []interface{}

	if len(cache.retMapping) == 0 {
		goto CacheNoHooks
	}

	lastID, err = result.LastInsertId()
	if err != nil {
		return ErrSyncFail
	}

	o.ID = int(lastID)
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == coincheckHistoryMapping["id"] {
		goto CacheNoHooks
	}

	uniqueMap, err = queries.BindMapping(coincheckHistoryType, coincheckHistoryMapping, nzUniques)
	if err != nil {
		return errors.Wrap(err, "models: unable to retrieve unique values for coincheck_history")
	}
	nzUniqueCols = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), uniqueMap)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, nzUniqueCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, nzUniqueCols...).Scan(returns...)
	if err != nil {
		return errors.Wrap(err, "models: unable to populate default values for coincheck_history")
	}

CacheNoHooks:
	if !cached {
		coincheckHistoryUpsertCacheMut.Lock()
		coincheckHistoryUpsertCache[key] = cache
		coincheckHistoryUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single CoincheckHistory record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *CoincheckHistory) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no CoincheckHistory provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), coincheckHistoryPrimaryKeyMapping)
	sql := "DELETE FROM `coincheck_history` WHERE `id`=?"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from coincheck_history")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for coincheck_history")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q coincheckHistoryQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no coincheckHistoryQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from coincheck_history")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for coincheck_history")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o CoincheckHistorySlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(coincheckHistoryBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), coincheckHistoryPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM `coincheck_history` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, coincheckHistoryPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from coincheckHistory slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for coincheck_history")
	}

	if len(coincheckHistoryAfterDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	return rowsAff, nil
}

// Reload refetches the object from the database
// using the primary keys with an executor.
func (o *CoincheckHistory) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindCoincheckHistory(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *CoincheckHistorySlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := CoincheckHistorySlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), coincheckHistoryPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT `coincheck_history`.* FROM `coincheck_history` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, coincheckHistoryPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in CoincheckHistorySlice")
	}

	*o = slice

	return nil
}

// CoincheckHistoryExists checks if the CoincheckHistory row exists.
func CoincheckHistoryExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from `coincheck_history` where `id`=? limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if coincheck_history exists")
	}

	return exists, nil
}
