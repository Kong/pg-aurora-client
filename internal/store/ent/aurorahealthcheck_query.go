// Code generated by ent, DO NOT EDIT.

package ent

import (
	"context"
	"fmt"
	"math"

	"entgo.io/ent/dialect/sql"
	"entgo.io/ent/dialect/sql/sqlgraph"
	"entgo.io/ent/schema/field"
	"github.com/kong/pg-aurora-client/internal/store/ent/aurorahealthcheck"
	"github.com/kong/pg-aurora-client/internal/store/ent/predicate"
)

// AuroraHealthCheckQuery is the builder for querying AuroraHealthCheck entities.
type AuroraHealthCheckQuery struct {
	config
	ctx        *QueryContext
	order      []OrderFunc
	inters     []Interceptor
	predicates []predicate.AuroraHealthCheck
	// intermediate query (i.e. traversal path).
	sql  *sql.Selector
	path func(context.Context) (*sql.Selector, error)
}

// Where adds a new predicate for the AuroraHealthCheckQuery builder.
func (ahcq *AuroraHealthCheckQuery) Where(ps ...predicate.AuroraHealthCheck) *AuroraHealthCheckQuery {
	ahcq.predicates = append(ahcq.predicates, ps...)
	return ahcq
}

// Limit the number of records to be returned by this query.
func (ahcq *AuroraHealthCheckQuery) Limit(limit int) *AuroraHealthCheckQuery {
	ahcq.ctx.Limit = &limit
	return ahcq
}

// Offset to start from.
func (ahcq *AuroraHealthCheckQuery) Offset(offset int) *AuroraHealthCheckQuery {
	ahcq.ctx.Offset = &offset
	return ahcq
}

// Unique configures the query builder to filter duplicate records on query.
// By default, unique is set to true, and can be disabled using this method.
func (ahcq *AuroraHealthCheckQuery) Unique(unique bool) *AuroraHealthCheckQuery {
	ahcq.ctx.Unique = &unique
	return ahcq
}

// Order specifies how the records should be ordered.
func (ahcq *AuroraHealthCheckQuery) Order(o ...OrderFunc) *AuroraHealthCheckQuery {
	ahcq.order = append(ahcq.order, o...)
	return ahcq
}

// First returns the first AuroraHealthCheck entity from the query.
// Returns a *NotFoundError when no AuroraHealthCheck was found.
func (ahcq *AuroraHealthCheckQuery) First(ctx context.Context) (*AuroraHealthCheck, error) {
	nodes, err := ahcq.Limit(1).All(setContextOp(ctx, ahcq.ctx, "First"))
	if err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nil, &NotFoundError{aurorahealthcheck.Label}
	}
	return nodes[0], nil
}

// FirstX is like First, but panics if an error occurs.
func (ahcq *AuroraHealthCheckQuery) FirstX(ctx context.Context) *AuroraHealthCheck {
	node, err := ahcq.First(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return node
}

// FirstID returns the first AuroraHealthCheck ID from the query.
// Returns a *NotFoundError when no AuroraHealthCheck ID was found.
func (ahcq *AuroraHealthCheckQuery) FirstID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = ahcq.Limit(1).IDs(setContextOp(ctx, ahcq.ctx, "FirstID")); err != nil {
		return
	}
	if len(ids) == 0 {
		err = &NotFoundError{aurorahealthcheck.Label}
		return
	}
	return ids[0], nil
}

// FirstIDX is like FirstID, but panics if an error occurs.
func (ahcq *AuroraHealthCheckQuery) FirstIDX(ctx context.Context) int {
	id, err := ahcq.FirstID(ctx)
	if err != nil && !IsNotFound(err) {
		panic(err)
	}
	return id
}

// Only returns a single AuroraHealthCheck entity found by the query, ensuring it only returns one.
// Returns a *NotSingularError when more than one AuroraHealthCheck entity is found.
// Returns a *NotFoundError when no AuroraHealthCheck entities are found.
func (ahcq *AuroraHealthCheckQuery) Only(ctx context.Context) (*AuroraHealthCheck, error) {
	nodes, err := ahcq.Limit(2).All(setContextOp(ctx, ahcq.ctx, "Only"))
	if err != nil {
		return nil, err
	}
	switch len(nodes) {
	case 1:
		return nodes[0], nil
	case 0:
		return nil, &NotFoundError{aurorahealthcheck.Label}
	default:
		return nil, &NotSingularError{aurorahealthcheck.Label}
	}
}

// OnlyX is like Only, but panics if an error occurs.
func (ahcq *AuroraHealthCheckQuery) OnlyX(ctx context.Context) *AuroraHealthCheck {
	node, err := ahcq.Only(ctx)
	if err != nil {
		panic(err)
	}
	return node
}

// OnlyID is like Only, but returns the only AuroraHealthCheck ID in the query.
// Returns a *NotSingularError when more than one AuroraHealthCheck ID is found.
// Returns a *NotFoundError when no entities are found.
func (ahcq *AuroraHealthCheckQuery) OnlyID(ctx context.Context) (id int, err error) {
	var ids []int
	if ids, err = ahcq.Limit(2).IDs(setContextOp(ctx, ahcq.ctx, "OnlyID")); err != nil {
		return
	}
	switch len(ids) {
	case 1:
		id = ids[0]
	case 0:
		err = &NotFoundError{aurorahealthcheck.Label}
	default:
		err = &NotSingularError{aurorahealthcheck.Label}
	}
	return
}

// OnlyIDX is like OnlyID, but panics if an error occurs.
func (ahcq *AuroraHealthCheckQuery) OnlyIDX(ctx context.Context) int {
	id, err := ahcq.OnlyID(ctx)
	if err != nil {
		panic(err)
	}
	return id
}

// All executes the query and returns a list of AuroraHealthChecks.
func (ahcq *AuroraHealthCheckQuery) All(ctx context.Context) ([]*AuroraHealthCheck, error) {
	ctx = setContextOp(ctx, ahcq.ctx, "All")
	if err := ahcq.prepareQuery(ctx); err != nil {
		return nil, err
	}
	qr := querierAll[[]*AuroraHealthCheck, *AuroraHealthCheckQuery]()
	return withInterceptors[[]*AuroraHealthCheck](ctx, ahcq, qr, ahcq.inters)
}

// AllX is like All, but panics if an error occurs.
func (ahcq *AuroraHealthCheckQuery) AllX(ctx context.Context) []*AuroraHealthCheck {
	nodes, err := ahcq.All(ctx)
	if err != nil {
		panic(err)
	}
	return nodes
}

// IDs executes the query and returns a list of AuroraHealthCheck IDs.
func (ahcq *AuroraHealthCheckQuery) IDs(ctx context.Context) (ids []int, err error) {
	if ahcq.ctx.Unique == nil && ahcq.path != nil {
		ahcq.Unique(true)
	}
	ctx = setContextOp(ctx, ahcq.ctx, "IDs")
	if err = ahcq.Select(aurorahealthcheck.FieldID).Scan(ctx, &ids); err != nil {
		return nil, err
	}
	return ids, nil
}

// IDsX is like IDs, but panics if an error occurs.
func (ahcq *AuroraHealthCheckQuery) IDsX(ctx context.Context) []int {
	ids, err := ahcq.IDs(ctx)
	if err != nil {
		panic(err)
	}
	return ids
}

// Count returns the count of the given query.
func (ahcq *AuroraHealthCheckQuery) Count(ctx context.Context) (int, error) {
	ctx = setContextOp(ctx, ahcq.ctx, "Count")
	if err := ahcq.prepareQuery(ctx); err != nil {
		return 0, err
	}
	return withInterceptors[int](ctx, ahcq, querierCount[*AuroraHealthCheckQuery](), ahcq.inters)
}

// CountX is like Count, but panics if an error occurs.
func (ahcq *AuroraHealthCheckQuery) CountX(ctx context.Context) int {
	count, err := ahcq.Count(ctx)
	if err != nil {
		panic(err)
	}
	return count
}

// Exist returns true if the query has elements in the graph.
func (ahcq *AuroraHealthCheckQuery) Exist(ctx context.Context) (bool, error) {
	ctx = setContextOp(ctx, ahcq.ctx, "Exist")
	switch _, err := ahcq.FirstID(ctx); {
	case IsNotFound(err):
		return false, nil
	case err != nil:
		return false, fmt.Errorf("ent: check existence: %w", err)
	default:
		return true, nil
	}
}

// ExistX is like Exist, but panics if an error occurs.
func (ahcq *AuroraHealthCheckQuery) ExistX(ctx context.Context) bool {
	exist, err := ahcq.Exist(ctx)
	if err != nil {
		panic(err)
	}
	return exist
}

// Clone returns a duplicate of the AuroraHealthCheckQuery builder, including all associated steps. It can be
// used to prepare common query builders and use them differently after the clone is made.
func (ahcq *AuroraHealthCheckQuery) Clone() *AuroraHealthCheckQuery {
	if ahcq == nil {
		return nil
	}
	return &AuroraHealthCheckQuery{
		config:     ahcq.config,
		ctx:        ahcq.ctx.Clone(),
		order:      append([]OrderFunc{}, ahcq.order...),
		inters:     append([]Interceptor{}, ahcq.inters...),
		predicates: append([]predicate.AuroraHealthCheck{}, ahcq.predicates...),
		// clone intermediate query.
		sql:  ahcq.sql.Clone(),
		path: ahcq.path,
	}
}

// GroupBy is used to group vertices by one or more fields/columns.
// It is often used with aggregate functions, like: count, max, mean, min, sum.
//
// Example:
//
//	var v []struct {
//		Ts time.Time `json:"ts,omitempty"`
//		Count int `json:"count,omitempty"`
//	}
//
//	client.AuroraHealthCheck.Query().
//		GroupBy(aurorahealthcheck.FieldTs).
//		Aggregate(ent.Count()).
//		Scan(ctx, &v)
func (ahcq *AuroraHealthCheckQuery) GroupBy(field string, fields ...string) *AuroraHealthCheckGroupBy {
	ahcq.ctx.Fields = append([]string{field}, fields...)
	grbuild := &AuroraHealthCheckGroupBy{build: ahcq}
	grbuild.flds = &ahcq.ctx.Fields
	grbuild.label = aurorahealthcheck.Label
	grbuild.scan = grbuild.Scan
	return grbuild
}

// Select allows the selection one or more fields/columns for the given query,
// instead of selecting all fields in the entity.
//
// Example:
//
//	var v []struct {
//		Ts time.Time `json:"ts,omitempty"`
//	}
//
//	client.AuroraHealthCheck.Query().
//		Select(aurorahealthcheck.FieldTs).
//		Scan(ctx, &v)
func (ahcq *AuroraHealthCheckQuery) Select(fields ...string) *AuroraHealthCheckSelect {
	ahcq.ctx.Fields = append(ahcq.ctx.Fields, fields...)
	sbuild := &AuroraHealthCheckSelect{AuroraHealthCheckQuery: ahcq}
	sbuild.label = aurorahealthcheck.Label
	sbuild.flds, sbuild.scan = &ahcq.ctx.Fields, sbuild.Scan
	return sbuild
}

// Aggregate returns a AuroraHealthCheckSelect configured with the given aggregations.
func (ahcq *AuroraHealthCheckQuery) Aggregate(fns ...AggregateFunc) *AuroraHealthCheckSelect {
	return ahcq.Select().Aggregate(fns...)
}

func (ahcq *AuroraHealthCheckQuery) prepareQuery(ctx context.Context) error {
	for _, inter := range ahcq.inters {
		if inter == nil {
			return fmt.Errorf("ent: uninitialized interceptor (forgotten import ent/runtime?)")
		}
		if trv, ok := inter.(Traverser); ok {
			if err := trv.Traverse(ctx, ahcq); err != nil {
				return err
			}
		}
	}
	for _, f := range ahcq.ctx.Fields {
		if !aurorahealthcheck.ValidColumn(f) {
			return &ValidationError{Name: f, err: fmt.Errorf("ent: invalid field %q for query", f)}
		}
	}
	if ahcq.path != nil {
		prev, err := ahcq.path(ctx)
		if err != nil {
			return err
		}
		ahcq.sql = prev
	}
	return nil
}

func (ahcq *AuroraHealthCheckQuery) sqlAll(ctx context.Context, hooks ...queryHook) ([]*AuroraHealthCheck, error) {
	var (
		nodes = []*AuroraHealthCheck{}
		_spec = ahcq.querySpec()
	)
	_spec.ScanValues = func(columns []string) ([]any, error) {
		return (*AuroraHealthCheck).scanValues(nil, columns)
	}
	_spec.Assign = func(columns []string, values []any) error {
		node := &AuroraHealthCheck{config: ahcq.config}
		nodes = append(nodes, node)
		return node.assignValues(columns, values)
	}
	for i := range hooks {
		hooks[i](ctx, _spec)
	}
	if err := sqlgraph.QueryNodes(ctx, ahcq.driver, _spec); err != nil {
		return nil, err
	}
	if len(nodes) == 0 {
		return nodes, nil
	}
	return nodes, nil
}

func (ahcq *AuroraHealthCheckQuery) sqlCount(ctx context.Context) (int, error) {
	_spec := ahcq.querySpec()
	_spec.Node.Columns = ahcq.ctx.Fields
	if len(ahcq.ctx.Fields) > 0 {
		_spec.Unique = ahcq.ctx.Unique != nil && *ahcq.ctx.Unique
	}
	return sqlgraph.CountNodes(ctx, ahcq.driver, _spec)
}

func (ahcq *AuroraHealthCheckQuery) querySpec() *sqlgraph.QuerySpec {
	_spec := sqlgraph.NewQuerySpec(aurorahealthcheck.Table, aurorahealthcheck.Columns, sqlgraph.NewFieldSpec(aurorahealthcheck.FieldID, field.TypeInt))
	_spec.From = ahcq.sql
	if unique := ahcq.ctx.Unique; unique != nil {
		_spec.Unique = *unique
	} else if ahcq.path != nil {
		_spec.Unique = true
	}
	if fields := ahcq.ctx.Fields; len(fields) > 0 {
		_spec.Node.Columns = make([]string, 0, len(fields))
		_spec.Node.Columns = append(_spec.Node.Columns, aurorahealthcheck.FieldID)
		for i := range fields {
			if fields[i] != aurorahealthcheck.FieldID {
				_spec.Node.Columns = append(_spec.Node.Columns, fields[i])
			}
		}
	}
	if ps := ahcq.predicates; len(ps) > 0 {
		_spec.Predicate = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	if limit := ahcq.ctx.Limit; limit != nil {
		_spec.Limit = *limit
	}
	if offset := ahcq.ctx.Offset; offset != nil {
		_spec.Offset = *offset
	}
	if ps := ahcq.order; len(ps) > 0 {
		_spec.Order = func(selector *sql.Selector) {
			for i := range ps {
				ps[i](selector)
			}
		}
	}
	return _spec
}

func (ahcq *AuroraHealthCheckQuery) sqlQuery(ctx context.Context) *sql.Selector {
	builder := sql.Dialect(ahcq.driver.Dialect())
	t1 := builder.Table(aurorahealthcheck.Table)
	columns := ahcq.ctx.Fields
	if len(columns) == 0 {
		columns = aurorahealthcheck.Columns
	}
	selector := builder.Select(t1.Columns(columns...)...).From(t1)
	if ahcq.sql != nil {
		selector = ahcq.sql
		selector.Select(selector.Columns(columns...)...)
	}
	if ahcq.ctx.Unique != nil && *ahcq.ctx.Unique {
		selector.Distinct()
	}
	for _, p := range ahcq.predicates {
		p(selector)
	}
	for _, p := range ahcq.order {
		p(selector)
	}
	if offset := ahcq.ctx.Offset; offset != nil {
		// limit is mandatory for offset clause. We start
		// with default value, and override it below if needed.
		selector.Offset(*offset).Limit(math.MaxInt32)
	}
	if limit := ahcq.ctx.Limit; limit != nil {
		selector.Limit(*limit)
	}
	return selector
}

// AuroraHealthCheckGroupBy is the group-by builder for AuroraHealthCheck entities.
type AuroraHealthCheckGroupBy struct {
	selector
	build *AuroraHealthCheckQuery
}

// Aggregate adds the given aggregation functions to the group-by query.
func (ahcgb *AuroraHealthCheckGroupBy) Aggregate(fns ...AggregateFunc) *AuroraHealthCheckGroupBy {
	ahcgb.fns = append(ahcgb.fns, fns...)
	return ahcgb
}

// Scan applies the selector query and scans the result into the given value.
func (ahcgb *AuroraHealthCheckGroupBy) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ahcgb.build.ctx, "GroupBy")
	if err := ahcgb.build.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*AuroraHealthCheckQuery, *AuroraHealthCheckGroupBy](ctx, ahcgb.build, ahcgb, ahcgb.build.inters, v)
}

func (ahcgb *AuroraHealthCheckGroupBy) sqlScan(ctx context.Context, root *AuroraHealthCheckQuery, v any) error {
	selector := root.sqlQuery(ctx).Select()
	aggregation := make([]string, 0, len(ahcgb.fns))
	for _, fn := range ahcgb.fns {
		aggregation = append(aggregation, fn(selector))
	}
	if len(selector.SelectedColumns()) == 0 {
		columns := make([]string, 0, len(*ahcgb.flds)+len(ahcgb.fns))
		for _, f := range *ahcgb.flds {
			columns = append(columns, selector.C(f))
		}
		columns = append(columns, aggregation...)
		selector.Select(columns...)
	}
	selector.GroupBy(selector.Columns(*ahcgb.flds...)...)
	if err := selector.Err(); err != nil {
		return err
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ahcgb.build.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}

// AuroraHealthCheckSelect is the builder for selecting fields of AuroraHealthCheck entities.
type AuroraHealthCheckSelect struct {
	*AuroraHealthCheckQuery
	selector
}

// Aggregate adds the given aggregation functions to the selector query.
func (ahcs *AuroraHealthCheckSelect) Aggregate(fns ...AggregateFunc) *AuroraHealthCheckSelect {
	ahcs.fns = append(ahcs.fns, fns...)
	return ahcs
}

// Scan applies the selector query and scans the result into the given value.
func (ahcs *AuroraHealthCheckSelect) Scan(ctx context.Context, v any) error {
	ctx = setContextOp(ctx, ahcs.ctx, "Select")
	if err := ahcs.prepareQuery(ctx); err != nil {
		return err
	}
	return scanWithInterceptors[*AuroraHealthCheckQuery, *AuroraHealthCheckSelect](ctx, ahcs.AuroraHealthCheckQuery, ahcs, ahcs.inters, v)
}

func (ahcs *AuroraHealthCheckSelect) sqlScan(ctx context.Context, root *AuroraHealthCheckQuery, v any) error {
	selector := root.sqlQuery(ctx)
	aggregation := make([]string, 0, len(ahcs.fns))
	for _, fn := range ahcs.fns {
		aggregation = append(aggregation, fn(selector))
	}
	switch n := len(*ahcs.selector.flds); {
	case n == 0 && len(aggregation) > 0:
		selector.Select(aggregation...)
	case n != 0 && len(aggregation) > 0:
		selector.AppendSelect(aggregation...)
	}
	rows := &sql.Rows{}
	query, args := selector.Query()
	if err := ahcs.driver.Query(ctx, query, args, rows); err != nil {
		return err
	}
	defer rows.Close()
	return sql.ScanSlice(rows, v)
}
