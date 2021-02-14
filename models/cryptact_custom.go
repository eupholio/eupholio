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
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/queries/qmhelper"
	"github.com/volatiletech/sqlboiler/v4/types"
	"github.com/volatiletech/strmangle"
)

// CryptactCustom is an object representing the database table.
type CryptactCustom struct {
	ID        int               `boil:"id" json:"id" toml:"id" yaml:"id"`
	Timestamp time.Time         `boil:"timestamp" json:"timestamp" toml:"timestamp" yaml:"timestamp"`
	Action    string            `boil:"action" json:"action" toml:"action" yaml:"action"`
	Source    string            `boil:"source" json:"source" toml:"source" yaml:"source"`
	Base      string            `boil:"base" json:"base" toml:"base" yaml:"base"`
	Volume    types.Decimal     `boil:"volume" json:"volume" toml:"volume" yaml:"volume"`
	Price     types.NullDecimal `boil:"price" json:"price,omitempty" toml:"price" yaml:"price,omitempty"`
	Counter   string            `boil:"counter" json:"counter" toml:"counter" yaml:"counter"`
	Fee       types.Decimal     `boil:"fee" json:"fee" toml:"fee" yaml:"fee"`
	FeeCcy    string            `boil:"fee_ccy" json:"fee_ccy" toml:"fee_ccy" yaml:"fee_ccy"`

	R *cryptactCustomR `boil:"-" json:"-" toml:"-" yaml:"-"`
	L cryptactCustomL  `boil:"-" json:"-" toml:"-" yaml:"-"`
}

var CryptactCustomColumns = struct {
	ID        string
	Timestamp string
	Action    string
	Source    string
	Base      string
	Volume    string
	Price     string
	Counter   string
	Fee       string
	FeeCcy    string
}{
	ID:        "id",
	Timestamp: "timestamp",
	Action:    "action",
	Source:    "source",
	Base:      "base",
	Volume:    "volume",
	Price:     "price",
	Counter:   "counter",
	Fee:       "fee",
	FeeCcy:    "fee_ccy",
}

// Generated where

var CryptactCustomWhere = struct {
	ID        whereHelperint
	Timestamp whereHelpertime_Time
	Action    whereHelperstring
	Source    whereHelperstring
	Base      whereHelperstring
	Volume    whereHelpertypes_Decimal
	Price     whereHelpertypes_NullDecimal
	Counter   whereHelperstring
	Fee       whereHelpertypes_Decimal
	FeeCcy    whereHelperstring
}{
	ID:        whereHelperint{field: "`cryptact_custom`.`id`"},
	Timestamp: whereHelpertime_Time{field: "`cryptact_custom`.`timestamp`"},
	Action:    whereHelperstring{field: "`cryptact_custom`.`action`"},
	Source:    whereHelperstring{field: "`cryptact_custom`.`source`"},
	Base:      whereHelperstring{field: "`cryptact_custom`.`base`"},
	Volume:    whereHelpertypes_Decimal{field: "`cryptact_custom`.`volume`"},
	Price:     whereHelpertypes_NullDecimal{field: "`cryptact_custom`.`price`"},
	Counter:   whereHelperstring{field: "`cryptact_custom`.`counter`"},
	Fee:       whereHelpertypes_Decimal{field: "`cryptact_custom`.`fee`"},
	FeeCcy:    whereHelperstring{field: "`cryptact_custom`.`fee_ccy`"},
}

// CryptactCustomRels is where relationship names are stored.
var CryptactCustomRels = struct {
}{}

// cryptactCustomR is where relationships are stored.
type cryptactCustomR struct {
}

// NewStruct creates a new relationship struct
func (*cryptactCustomR) NewStruct() *cryptactCustomR {
	return &cryptactCustomR{}
}

// cryptactCustomL is where Load methods for each relationship are stored.
type cryptactCustomL struct{}

var (
	cryptactCustomAllColumns            = []string{"id", "timestamp", "action", "source", "base", "volume", "price", "counter", "fee", "fee_ccy"}
	cryptactCustomColumnsWithoutDefault = []string{"timestamp", "action", "source", "base", "volume", "price", "counter", "fee", "fee_ccy"}
	cryptactCustomColumnsWithDefault    = []string{"id"}
	cryptactCustomPrimaryKeyColumns     = []string{"id"}
)

type (
	// CryptactCustomSlice is an alias for a slice of pointers to CryptactCustom.
	// This should generally be used opposed to []CryptactCustom.
	CryptactCustomSlice []*CryptactCustom
	// CryptactCustomHook is the signature for custom CryptactCustom hook methods
	CryptactCustomHook func(context.Context, boil.ContextExecutor, *CryptactCustom) error

	cryptactCustomQuery struct {
		*queries.Query
	}
)

// Cache for insert, update and upsert
var (
	cryptactCustomType                 = reflect.TypeOf(&CryptactCustom{})
	cryptactCustomMapping              = queries.MakeStructMapping(cryptactCustomType)
	cryptactCustomPrimaryKeyMapping, _ = queries.BindMapping(cryptactCustomType, cryptactCustomMapping, cryptactCustomPrimaryKeyColumns)
	cryptactCustomInsertCacheMut       sync.RWMutex
	cryptactCustomInsertCache          = make(map[string]insertCache)
	cryptactCustomUpdateCacheMut       sync.RWMutex
	cryptactCustomUpdateCache          = make(map[string]updateCache)
	cryptactCustomUpsertCacheMut       sync.RWMutex
	cryptactCustomUpsertCache          = make(map[string]insertCache)
)

var (
	// Force time package dependency for automated UpdatedAt/CreatedAt.
	_ = time.Second
	// Force qmhelper dependency for where clause generation (which doesn't
	// always happen)
	_ = qmhelper.Where
)

var cryptactCustomBeforeInsertHooks []CryptactCustomHook
var cryptactCustomBeforeUpdateHooks []CryptactCustomHook
var cryptactCustomBeforeDeleteHooks []CryptactCustomHook
var cryptactCustomBeforeUpsertHooks []CryptactCustomHook

var cryptactCustomAfterInsertHooks []CryptactCustomHook
var cryptactCustomAfterSelectHooks []CryptactCustomHook
var cryptactCustomAfterUpdateHooks []CryptactCustomHook
var cryptactCustomAfterDeleteHooks []CryptactCustomHook
var cryptactCustomAfterUpsertHooks []CryptactCustomHook

// doBeforeInsertHooks executes all "before insert" hooks.
func (o *CryptactCustom) doBeforeInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range cryptactCustomBeforeInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpdateHooks executes all "before Update" hooks.
func (o *CryptactCustom) doBeforeUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range cryptactCustomBeforeUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeDeleteHooks executes all "before Delete" hooks.
func (o *CryptactCustom) doBeforeDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range cryptactCustomBeforeDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doBeforeUpsertHooks executes all "before Upsert" hooks.
func (o *CryptactCustom) doBeforeUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range cryptactCustomBeforeUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterInsertHooks executes all "after Insert" hooks.
func (o *CryptactCustom) doAfterInsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range cryptactCustomAfterInsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterSelectHooks executes all "after Select" hooks.
func (o *CryptactCustom) doAfterSelectHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range cryptactCustomAfterSelectHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpdateHooks executes all "after Update" hooks.
func (o *CryptactCustom) doAfterUpdateHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range cryptactCustomAfterUpdateHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterDeleteHooks executes all "after Delete" hooks.
func (o *CryptactCustom) doAfterDeleteHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range cryptactCustomAfterDeleteHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// doAfterUpsertHooks executes all "after Upsert" hooks.
func (o *CryptactCustom) doAfterUpsertHooks(ctx context.Context, exec boil.ContextExecutor) (err error) {
	if boil.HooksAreSkipped(ctx) {
		return nil
	}

	for _, hook := range cryptactCustomAfterUpsertHooks {
		if err := hook(ctx, exec, o); err != nil {
			return err
		}
	}

	return nil
}

// AddCryptactCustomHook registers your hook function for all future operations.
func AddCryptactCustomHook(hookPoint boil.HookPoint, cryptactCustomHook CryptactCustomHook) {
	switch hookPoint {
	case boil.BeforeInsertHook:
		cryptactCustomBeforeInsertHooks = append(cryptactCustomBeforeInsertHooks, cryptactCustomHook)
	case boil.BeforeUpdateHook:
		cryptactCustomBeforeUpdateHooks = append(cryptactCustomBeforeUpdateHooks, cryptactCustomHook)
	case boil.BeforeDeleteHook:
		cryptactCustomBeforeDeleteHooks = append(cryptactCustomBeforeDeleteHooks, cryptactCustomHook)
	case boil.BeforeUpsertHook:
		cryptactCustomBeforeUpsertHooks = append(cryptactCustomBeforeUpsertHooks, cryptactCustomHook)
	case boil.AfterInsertHook:
		cryptactCustomAfterInsertHooks = append(cryptactCustomAfterInsertHooks, cryptactCustomHook)
	case boil.AfterSelectHook:
		cryptactCustomAfterSelectHooks = append(cryptactCustomAfterSelectHooks, cryptactCustomHook)
	case boil.AfterUpdateHook:
		cryptactCustomAfterUpdateHooks = append(cryptactCustomAfterUpdateHooks, cryptactCustomHook)
	case boil.AfterDeleteHook:
		cryptactCustomAfterDeleteHooks = append(cryptactCustomAfterDeleteHooks, cryptactCustomHook)
	case boil.AfterUpsertHook:
		cryptactCustomAfterUpsertHooks = append(cryptactCustomAfterUpsertHooks, cryptactCustomHook)
	}
}

// One returns a single cryptactCustom record from the query.
func (q cryptactCustomQuery) One(ctx context.Context, exec boil.ContextExecutor) (*CryptactCustom, error) {
	o := &CryptactCustom{}

	queries.SetLimit(q.Query, 1)

	err := q.Bind(ctx, exec, o)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: failed to execute a one query for cryptact_custom")
	}

	if err := o.doAfterSelectHooks(ctx, exec); err != nil {
		return o, err
	}

	return o, nil
}

// All returns all CryptactCustom records from the query.
func (q cryptactCustomQuery) All(ctx context.Context, exec boil.ContextExecutor) (CryptactCustomSlice, error) {
	var o []*CryptactCustom

	err := q.Bind(ctx, exec, &o)
	if err != nil {
		return nil, errors.Wrap(err, "models: failed to assign all query results to CryptactCustom slice")
	}

	if len(cryptactCustomAfterSelectHooks) != 0 {
		for _, obj := range o {
			if err := obj.doAfterSelectHooks(ctx, exec); err != nil {
				return o, err
			}
		}
	}

	return o, nil
}

// Count returns the count of all CryptactCustom records in the query.
func (q cryptactCustomQuery) Count(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to count cryptact_custom rows")
	}

	return count, nil
}

// Exists checks if the row exists in the table.
func (q cryptactCustomQuery) Exists(ctx context.Context, exec boil.ContextExecutor) (bool, error) {
	var count int64

	queries.SetSelect(q.Query, nil)
	queries.SetCount(q.Query)
	queries.SetLimit(q.Query, 1)

	err := q.Query.QueryRowContext(ctx, exec).Scan(&count)
	if err != nil {
		return false, errors.Wrap(err, "models: failed to check if cryptact_custom exists")
	}

	return count > 0, nil
}

// CryptactCustoms retrieves all the records using an executor.
func CryptactCustoms(mods ...qm.QueryMod) cryptactCustomQuery {
	mods = append(mods, qm.From("`cryptact_custom`"))
	return cryptactCustomQuery{NewQuery(mods...)}
}

// FindCryptactCustom retrieves a single record by ID with an executor.
// If selectCols is empty Find will return all columns.
func FindCryptactCustom(ctx context.Context, exec boil.ContextExecutor, iD int, selectCols ...string) (*CryptactCustom, error) {
	cryptactCustomObj := &CryptactCustom{}

	sel := "*"
	if len(selectCols) > 0 {
		sel = strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, selectCols), ",")
	}
	query := fmt.Sprintf(
		"select %s from `cryptact_custom` where `id`=?", sel,
	)

	q := queries.Raw(query, iD)

	err := q.Bind(ctx, exec, cryptactCustomObj)
	if err != nil {
		if errors.Cause(err) == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, errors.Wrap(err, "models: unable to select from cryptact_custom")
	}

	return cryptactCustomObj, nil
}

// Insert a single record using an executor.
// See boil.Columns.InsertColumnSet documentation to understand column list inference for inserts.
func (o *CryptactCustom) Insert(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) error {
	if o == nil {
		return errors.New("models: no cryptact_custom provided for insertion")
	}

	var err error

	if err := o.doBeforeInsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(cryptactCustomColumnsWithDefault, o)

	key := makeCacheKey(columns, nzDefaults)
	cryptactCustomInsertCacheMut.RLock()
	cache, cached := cryptactCustomInsertCache[key]
	cryptactCustomInsertCacheMut.RUnlock()

	if !cached {
		wl, returnColumns := columns.InsertColumnSet(
			cryptactCustomAllColumns,
			cryptactCustomColumnsWithDefault,
			cryptactCustomColumnsWithoutDefault,
			nzDefaults,
		)

		cache.valueMapping, err = queries.BindMapping(cryptactCustomType, cryptactCustomMapping, wl)
		if err != nil {
			return err
		}
		cache.retMapping, err = queries.BindMapping(cryptactCustomType, cryptactCustomMapping, returnColumns)
		if err != nil {
			return err
		}
		if len(wl) != 0 {
			cache.query = fmt.Sprintf("INSERT INTO `cryptact_custom` (`%s`) %%sVALUES (%s)%%s", strings.Join(wl, "`,`"), strmangle.Placeholders(dialect.UseIndexPlaceholders, len(wl), 1, 1))
		} else {
			cache.query = "INSERT INTO `cryptact_custom` () VALUES ()%s%s"
		}

		var queryOutput, queryReturning string

		if len(cache.retMapping) != 0 {
			cache.retQuery = fmt.Sprintf("SELECT `%s` FROM `cryptact_custom` WHERE %s", strings.Join(returnColumns, "`,`"), strmangle.WhereClause("`", "`", 0, cryptactCustomPrimaryKeyColumns))
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
		return errors.Wrap(err, "models: unable to insert into cryptact_custom")
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
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == cryptactCustomMapping["id"] {
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
		return errors.Wrap(err, "models: unable to populate default values for cryptact_custom")
	}

CacheNoHooks:
	if !cached {
		cryptactCustomInsertCacheMut.Lock()
		cryptactCustomInsertCache[key] = cache
		cryptactCustomInsertCacheMut.Unlock()
	}

	return o.doAfterInsertHooks(ctx, exec)
}

// Update uses an executor to update the CryptactCustom.
// See boil.Columns.UpdateColumnSet documentation to understand column list inference for updates.
// Update does not automatically update the record in case of default values. Use .Reload() to refresh the records.
func (o *CryptactCustom) Update(ctx context.Context, exec boil.ContextExecutor, columns boil.Columns) (int64, error) {
	var err error
	if err = o.doBeforeUpdateHooks(ctx, exec); err != nil {
		return 0, err
	}
	key := makeCacheKey(columns, nil)
	cryptactCustomUpdateCacheMut.RLock()
	cache, cached := cryptactCustomUpdateCache[key]
	cryptactCustomUpdateCacheMut.RUnlock()

	if !cached {
		wl := columns.UpdateColumnSet(
			cryptactCustomAllColumns,
			cryptactCustomPrimaryKeyColumns,
		)

		if !columns.IsWhitelist() {
			wl = strmangle.SetComplement(wl, []string{"created_at"})
		}
		if len(wl) == 0 {
			return 0, errors.New("models: unable to update cryptact_custom, could not build whitelist")
		}

		cache.query = fmt.Sprintf("UPDATE `cryptact_custom` SET %s WHERE %s",
			strmangle.SetParamNames("`", "`", 0, wl),
			strmangle.WhereClause("`", "`", 0, cryptactCustomPrimaryKeyColumns),
		)
		cache.valueMapping, err = queries.BindMapping(cryptactCustomType, cryptactCustomMapping, append(wl, cryptactCustomPrimaryKeyColumns...))
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
		return 0, errors.Wrap(err, "models: unable to update cryptact_custom row")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by update for cryptact_custom")
	}

	if !cached {
		cryptactCustomUpdateCacheMut.Lock()
		cryptactCustomUpdateCache[key] = cache
		cryptactCustomUpdateCacheMut.Unlock()
	}

	return rowsAff, o.doAfterUpdateHooks(ctx, exec)
}

// UpdateAll updates all rows with the specified column values.
func (q cryptactCustomQuery) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
	queries.SetUpdate(q.Query, cols)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all for cryptact_custom")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected for cryptact_custom")
	}

	return rowsAff, nil
}

// UpdateAll updates all rows with the specified column values, using an executor.
func (o CryptactCustomSlice) UpdateAll(ctx context.Context, exec boil.ContextExecutor, cols M) (int64, error) {
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
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), cryptactCustomPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := fmt.Sprintf("UPDATE `cryptact_custom` SET %s WHERE %s",
		strmangle.SetParamNames("`", "`", 0, colNames),
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, cryptactCustomPrimaryKeyColumns, len(o)))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to update all in cryptactCustom slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to retrieve rows affected all in update all cryptactCustom")
	}
	return rowsAff, nil
}

var mySQLCryptactCustomUniqueColumns = []string{
	"id",
}

// Upsert attempts an insert using an executor, and does an update or ignore on conflict.
// See boil.Columns documentation for how to properly use updateColumns and insertColumns.
func (o *CryptactCustom) Upsert(ctx context.Context, exec boil.ContextExecutor, updateColumns, insertColumns boil.Columns) error {
	if o == nil {
		return errors.New("models: no cryptact_custom provided for upsert")
	}

	if err := o.doBeforeUpsertHooks(ctx, exec); err != nil {
		return err
	}

	nzDefaults := queries.NonZeroDefaultSet(cryptactCustomColumnsWithDefault, o)
	nzUniques := queries.NonZeroDefaultSet(mySQLCryptactCustomUniqueColumns, o)

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

	cryptactCustomUpsertCacheMut.RLock()
	cache, cached := cryptactCustomUpsertCache[key]
	cryptactCustomUpsertCacheMut.RUnlock()

	var err error

	if !cached {
		insert, ret := insertColumns.InsertColumnSet(
			cryptactCustomAllColumns,
			cryptactCustomColumnsWithDefault,
			cryptactCustomColumnsWithoutDefault,
			nzDefaults,
		)
		update := updateColumns.UpdateColumnSet(
			cryptactCustomAllColumns,
			cryptactCustomPrimaryKeyColumns,
		)

		if !updateColumns.IsNone() && len(update) == 0 {
			return errors.New("models: unable to upsert cryptact_custom, could not build update column list")
		}

		ret = strmangle.SetComplement(ret, nzUniques)
		cache.query = buildUpsertQueryMySQL(dialect, "`cryptact_custom`", update, insert)
		cache.retQuery = fmt.Sprintf(
			"SELECT %s FROM `cryptact_custom` WHERE %s",
			strings.Join(strmangle.IdentQuoteSlice(dialect.LQ, dialect.RQ, ret), ","),
			strmangle.WhereClause("`", "`", 0, nzUniques),
		)

		cache.valueMapping, err = queries.BindMapping(cryptactCustomType, cryptactCustomMapping, insert)
		if err != nil {
			return err
		}
		if len(ret) != 0 {
			cache.retMapping, err = queries.BindMapping(cryptactCustomType, cryptactCustomMapping, ret)
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
		return errors.Wrap(err, "models: unable to upsert for cryptact_custom")
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
	if lastID != 0 && len(cache.retMapping) == 1 && cache.retMapping[0] == cryptactCustomMapping["id"] {
		goto CacheNoHooks
	}

	uniqueMap, err = queries.BindMapping(cryptactCustomType, cryptactCustomMapping, nzUniques)
	if err != nil {
		return errors.Wrap(err, "models: unable to retrieve unique values for cryptact_custom")
	}
	nzUniqueCols = queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), uniqueMap)

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, cache.retQuery)
		fmt.Fprintln(writer, nzUniqueCols...)
	}
	err = exec.QueryRowContext(ctx, cache.retQuery, nzUniqueCols...).Scan(returns...)
	if err != nil {
		return errors.Wrap(err, "models: unable to populate default values for cryptact_custom")
	}

CacheNoHooks:
	if !cached {
		cryptactCustomUpsertCacheMut.Lock()
		cryptactCustomUpsertCache[key] = cache
		cryptactCustomUpsertCacheMut.Unlock()
	}

	return o.doAfterUpsertHooks(ctx, exec)
}

// Delete deletes a single CryptactCustom record with an executor.
// Delete will match against the primary key column to find the record to delete.
func (o *CryptactCustom) Delete(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if o == nil {
		return 0, errors.New("models: no CryptactCustom provided for delete")
	}

	if err := o.doBeforeDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	args := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(o)), cryptactCustomPrimaryKeyMapping)
	sql := "DELETE FROM `cryptact_custom` WHERE `id`=?"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args...)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete from cryptact_custom")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by delete for cryptact_custom")
	}

	if err := o.doAfterDeleteHooks(ctx, exec); err != nil {
		return 0, err
	}

	return rowsAff, nil
}

// DeleteAll deletes all matching rows.
func (q cryptactCustomQuery) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if q.Query == nil {
		return 0, errors.New("models: no cryptactCustomQuery provided for delete all")
	}

	queries.SetDelete(q.Query)

	result, err := q.Query.ExecContext(ctx, exec)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from cryptact_custom")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for cryptact_custom")
	}

	return rowsAff, nil
}

// DeleteAll deletes all rows in the slice, using an executor.
func (o CryptactCustomSlice) DeleteAll(ctx context.Context, exec boil.ContextExecutor) (int64, error) {
	if len(o) == 0 {
		return 0, nil
	}

	if len(cryptactCustomBeforeDeleteHooks) != 0 {
		for _, obj := range o {
			if err := obj.doBeforeDeleteHooks(ctx, exec); err != nil {
				return 0, err
			}
		}
	}

	var args []interface{}
	for _, obj := range o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), cryptactCustomPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "DELETE FROM `cryptact_custom` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, cryptactCustomPrimaryKeyColumns, len(o))

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, args)
	}
	result, err := exec.ExecContext(ctx, sql, args...)
	if err != nil {
		return 0, errors.Wrap(err, "models: unable to delete all from cryptactCustom slice")
	}

	rowsAff, err := result.RowsAffected()
	if err != nil {
		return 0, errors.Wrap(err, "models: failed to get rows affected by deleteall for cryptact_custom")
	}

	if len(cryptactCustomAfterDeleteHooks) != 0 {
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
func (o *CryptactCustom) Reload(ctx context.Context, exec boil.ContextExecutor) error {
	ret, err := FindCryptactCustom(ctx, exec, o.ID)
	if err != nil {
		return err
	}

	*o = *ret
	return nil
}

// ReloadAll refetches every row with matching primary key column values
// and overwrites the original object slice with the newly updated slice.
func (o *CryptactCustomSlice) ReloadAll(ctx context.Context, exec boil.ContextExecutor) error {
	if o == nil || len(*o) == 0 {
		return nil
	}

	slice := CryptactCustomSlice{}
	var args []interface{}
	for _, obj := range *o {
		pkeyArgs := queries.ValuesFromMapping(reflect.Indirect(reflect.ValueOf(obj)), cryptactCustomPrimaryKeyMapping)
		args = append(args, pkeyArgs...)
	}

	sql := "SELECT `cryptact_custom`.* FROM `cryptact_custom` WHERE " +
		strmangle.WhereClauseRepeated(string(dialect.LQ), string(dialect.RQ), 0, cryptactCustomPrimaryKeyColumns, len(*o))

	q := queries.Raw(sql, args...)

	err := q.Bind(ctx, exec, &slice)
	if err != nil {
		return errors.Wrap(err, "models: unable to reload all in CryptactCustomSlice")
	}

	*o = slice

	return nil
}

// CryptactCustomExists checks if the CryptactCustom row exists.
func CryptactCustomExists(ctx context.Context, exec boil.ContextExecutor, iD int) (bool, error) {
	var exists bool
	sql := "select exists(select 1 from `cryptact_custom` where `id`=? limit 1)"

	if boil.IsDebug(ctx) {
		writer := boil.DebugWriterFrom(ctx)
		fmt.Fprintln(writer, sql)
		fmt.Fprintln(writer, iD)
	}
	row := exec.QueryRowContext(ctx, sql, iD)

	err := row.Scan(&exists)
	if err != nil {
		return false, errors.Wrap(err, "models: unable to check if cryptact_custom exists")
	}

	return exists, nil
}
