package orm

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

type SegmentEntity struct {
	ID      string
	Place   Place
	PlaceID int
	Note    string
}

type Place struct {
	ID   int
	City string
}

type Segment struct {
	Start   SegmentEntity
	StartID string
	End     SegmentEntity
	EndID   string
}

var _ = Describe("InsertCascade", func() {
	It("single reference", func() {
		q := NewQuery(nil, &SegmentEntity{Place: Place{City: "New York"}})

		s := insertCascadeQueryString(q)

		Expect(s).To(Equal(`WITH "places" AS (INSERT INTO "places" ("id", "city") VALUES (DEFAULT, 'New York') RETURNING "id") INSERT INTO "segment_entities" ("id", "place_id") VALUES (DEFAULT, (SELECT "place"."id" FROM "places" AS "place")) RETURNING "id", "place_id"`))
	})

	FIt("two references", func() {
		q := NewQuery(nil, &Segment{
			Start: SegmentEntity{
				Note: "start-segment-entity",
			},
			End: SegmentEntity{
				Note: "end-segment-entity",
			},
		})

		s := insertCascadeQueryString(q)

		Expect(s).To(Equal(`WITH "Start_segment_entities" AS (INSERT INTO "segment_entities" ("id", "place_id", "note") VALUES (DEFAULT, DEFAULT, 'start-segment-entity') RETURNING "id", "place_id"), "End_segment_entities" AS (INSERT INTO "segment_entities" ("id", "place_id", "note") VALUES (DEFAULT, DEFAULT, 'end-segment-entity') RETURNING "id", "place_id") INSERT INTO "segments" ("start_id", "end_id") VALUES ((SELECT "id" FROM "Start_segment_entities"), (SELECT "id" FROM "End_segment_entities")) RETURNING "start_id", "end_id"`))
	})
})

func insertCascadeQueryString(q *Query) string {
	ins := newInsertCascadeQuery(q)
	return queryString(ins)
}
