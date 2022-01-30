package tests

import (
	"context"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/paulmach/orb"
	"github.com/stretchr/testify/assert"
)

func TestGeoPoint(t *testing.T) {
	var (
		ctx       = context.Background()
		conn, err = clickhouse.Open(&clickhouse.Options{
			Addr: []string{"127.0.0.1:9000"},
			Auth: clickhouse.Auth{
				Database: "default",
				Username: "default",
				Password: "",
			},
			Compression: &clickhouse.Compression{
				Method: clickhouse.CompressionLZ4,
			},
			Settings: clickhouse.Settings{
				"allow_experimental_geo_types": 1,
			},
		})
	)
	if assert.NoError(t, err) {
		if err := checkMinServerVersion(conn, 21, 12); err != nil {
			t.Skip(err.Error())
			return
		}
		const ddl = `
		CREATE TEMPORARY TABLE test_geo_point (
			Col1 Point
			, Col2 Array(Point)
		)
		`
		if err := conn.Exec(ctx, ddl); assert.NoError(t, err) {
			if batch, err := conn.PrepareBatch(ctx, "INSERT INTO test_geo_point"); assert.NoError(t, err) {
				if err := batch.Append(
					orb.Point{11, 22},
					[]orb.Point{
						{1, 2},
						{3, 4},
					},
				); assert.NoError(t, err) {
					if assert.NoError(t, batch.Send()) {
						var (
							col1 orb.Point
							col2 []orb.Point
						)
						if err := conn.QueryRow(ctx, "SELECT * FROM test_geo_point").Scan(&col1, &col2); assert.NoError(t, err) {
							assert.Equal(t, orb.Point{11, 22}, col1)
							assert.Equal(t, []orb.Point{
								{1, 2},
								{3, 4},
							}, col2)
						}
					}
				}
			}
		}
	}
}
