package orm

import (
	"reflect"
)

func InsertCascade(db DB, model ...interface{}) error {
	_, err := NewQuery(db, model...).InsertCascade()
	return err
}

type insertCascadeQuery struct {
	q               *Query
	returningFields []*Field
	placeholder     bool
}

var _ queryCommand = (*insertCascadeQuery)(nil)

func newInsertCascadeQuery(q *Query) *insertCascadeQuery {
	return &insertCascadeQuery{
		q: q,
	}
}

func (q *insertCascadeQuery) Operation() string {
	return InsertOp
}

func (q *insertCascadeQuery) Clone() queryCommand {
	return &insertCascadeQuery{
		q:           q.q.Clone(),
		placeholder: q.placeholder,
	}
}

func (q *insertCascadeQuery) Query() *Query {
	return q.q
}

var _ TemplateAppender = (*insertCascadeQuery)(nil)

func (q *insertCascadeQuery) AppendTemplate(b []byte) ([]byte, error) {
	cp := q.Clone().(*insertCascadeQuery)
	cp.placeholder = true
	return cp.AppendQuery(dummyFormatter{}, b)
}

var _ QueryAppender = (*insertCascadeQuery)(nil)

func (q *insertCascadeQuery) AppendQuery(fmter QueryFormatter, b []byte) (_ []byte, err error) {
	table := q.q.model.Table()

	withQuery := q.q
	for fieldName, rel := range table.Relations {
		withQuery = q.withRelation(fmter, b, rel, q.q.model.Root(), fieldName, withQuery)
	}

	for fieldName, rel := range table.Relations {
		relatedTableAlias := fieldName + "_" + rel.JoinTable.Name
		s := newSelectQuery(NewQuery(nil).Column("id").Table(relatedTableAlias))
		ss, _ := s.AppendQuery(fmter, nil)
		withQuery.Value(rel.FKs[0].SQLName, "("+string(ss)+")")
	}

	insert := newInsertQuery(withQuery)
	b, err = insert.AppendQuery(fmter, b)

	return b, q.q.stickyErr
}

func (q insertCascadeQuery) withRelation(fmter QueryFormatter, b []byte, rel *Relation, root reflect.Value, fieldName string, withQuery *Query) *Query {

	relatedStruct := root.FieldByName(fieldName)
	relatedStructQuery := NewQuery(nil, newStructTableModelValue(relatedStruct))

	// "<fieldName>_related_table"
	relatedTableAlias := fieldName + "_" + rel.JoinTable.Name

	// s := newSelectQuery(NewQuery(nil).Column("id").Table(relatedTableAlias))
	// ss, _ := s.AppendQuery(fmter, nil)

	// select id from related_table
	//sq := newSelectQuery(relatedStructQuery.ColumnExpr("?TablePKs"))
	//selectQuery, _ := sq.AppendQuery(NewFormatter().WithModel(sq), nil)

	// WITH "<fieldName>_related_table" AS (INSERT INTO "related_table" ("id") VALUES (DEFAULT) RETURNING "id")
	return withQuery.WithInsert(relatedTableAlias, relatedStructQuery)

	// insert := newInsertQuery(withQuery.Value(rel.FKs[0].SQLName, "("+string(ss)+")"))

	// b, err = insert.AppendQuery(fmter, b)

	// fmt.Println(string(b))
	// return b, nil
}
