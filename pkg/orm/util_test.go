package orm

import (
	"database/sql"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestToString(t *testing.T) {
	assert := assert.New(t)
	assert.Equal(ToString(nil), "", "should have been empty string")
	assert.Equal(ToString(""), "", "should have been empty string")
	assert.Equal(ToString("123"), "123", "should have been 123")
	assert.Equal(ToString(sql.NullString{String: "", Valid: false}), "", "should have been empty string")
	assert.Equal(ToString(&sql.NullString{String: "", Valid: false}), "", "should have been empty string")
	assert.Equal(ToString(sql.NullString{String: "yes", Valid: true}), "yes", "should have been yes")
	assert.Equal(ToString(&sql.NullString{String: "yes", Valid: true}), "yes", "should have been yes")
	assert.Equal(ToString(123), "123", "should have been 123")
	assert.Equal(ToString(int32(123)), "123", "should have been 123")
	assert.Equal(ToString(int64(123)), "123", "should have been 123")
	assert.Equal(ToString(float32(123)), "123.000000", "should have been 123.000000")
	assert.Equal(ToString(float64(123)), "123.000000", "should have been 123.000000")
	assert.Equal(ToString(true), "true", "should have been true")
	assert.Equal(ToString(false), "false", "should have been false")
	assert.Equal(ToString(sql.NullInt64{Int64: 123, Valid: true}), "123", "should have been 123")
	assert.Equal(ToString(&sql.NullInt64{Int64: 123, Valid: true}), "123", "should have been 123")
	assert.Equal(ToString(&sql.NullInt64{Int64: 0, Valid: false}), "0", "should have been 0")
	assert.Equal(ToString(&sql.NullFloat64{Float64: 0, Valid: false}), "0.000000", "should have been 0.000000")
	assert.Equal(ToString(&sql.NullFloat64{Float64: 123, Valid: true}), "123.000000", "should have been 123.000000")
	assert.Equal(ToString(sql.NullFloat64{Float64: 123, Valid: true}), "123.000000", "should have been 123.000000")
	assert.Equal(ToString(sql.NullBool{Bool: true, Valid: true}), "true", "should have been true")
	assert.Equal(ToString(&sql.NullBool{Bool: true, Valid: true}), "true", "should have been true")
	assert.Equal(ToString(sql.NullBool{Bool: false, Valid: true}), "false", "should have been false")
	assert.Equal(ToString(&sql.NullBool{Bool: false, Valid: true}), "false", "should have been false")
	tv := time.Now()
	tvs := tv.String()
	assert.Equal(ToString(tv), tvs, "should have been "+tvs)
	assert.Equal(ToString(mysql.NullTime{Time: tv, Valid: true}), tvs, "should have been "+tvs)
}

func TestSQLString(t *testing.T) {
	assert := assert.New(t)
	s := "select * from foo"
	str := ToSQLString(s)
	assert.Equal(str.String, s, "should have been "+s)
}

func TestSQLDate(t *testing.T) {
	assert := assert.New(t)
	v := ToSQLDate(nil)
	assert.Equal(v.Valid, false, "should have been invalid")
	v = ToSQLDate("now")
	assert.Equal(v.Valid, true, "should have been valid")
	tv := time.Now()
	v = ToSQLDate(tv)
	assert.Equal(v.Valid, true, "should have been valid")
	v = ToSQLDate("2017-03-17T21:35:27Z")
	assert.Equal(v.Valid, true, "should have been valid")
	v = ToSQLDate(ISODate())
	assert.Equal(v.Valid, true, "should have been valid")
}

func TestInt64(t *testing.T) {
	assert := assert.New(t)
	v := toInt64("123")
	assert.Equal(v, int64(123), "should have been 123")
	v = toInt64("")
	assert.Equal(v, int64(0), "should have been 0")
}

func TestInt32(t *testing.T) {
	assert := assert.New(t)
	v := toInt32("123")
	assert.Equal(v, int32(123), "should have been 123")
	v = toInt32("")
	assert.Equal(v, int32(0), "should have been 0")
}

func TestFloat64(t *testing.T) {
	assert := assert.New(t)
	v := toFloat64("123.0")
	assert.Equal(v, float64(123), "should have been 123.0000")
	v = toFloat64("")
	assert.Equal(v, float64(0), "should have been 0.0000")
}

func TestSQLInt64(t *testing.T) {
	assert := assert.New(t)
	v := ToSQLInt64("123")
	assert.Equal(v.Valid, true, "should have been true")
	assert.Equal(v.Int64, int64(123), "should have been 123")

	v = ToSQLInt64("")
	assert.Equal(v.Valid, false, "should have been false")
	assert.Equal(v.Int64, int64(0), "should have been 0")
}

func TestSQLFloat64(t *testing.T) {
	assert := assert.New(t)
	v := ToSQLFloat64("123.0")
	assert.Equal(v.Valid, true, "should have been true")
	assert.Equal(v.Float64, float64(123.0), "should have been 123.0000")

	v = ToSQLFloat64("")
	assert.Equal(v.Valid, false, "should have been false")
	assert.Equal(v.Float64, float64(0), "should have been 0")
}

func TestSQLBool(t *testing.T) {
	assert := assert.New(t)
	v := ToSQLBool("true")
	assert.Equal(v.Valid, true, "should have been true")
	assert.Equal(v.Bool, true, "should have been true")
	v = ToSQLBool("false")
	assert.Equal(v.Valid, true, "should have been true")
	assert.Equal(v.Bool, false, "should have been false")
	v = ToSQLBool("")
	assert.Equal(v.Valid, false, "should have been false")
	assert.Equal(v.Bool, false, "should have been false")
	v = ToSQLBool(1)
	assert.Equal(v.Valid, true, "should have been true")
	assert.Equal(v.Bool, true, "should have been true")
	v = ToSQLBool(0)
	assert.Equal(v.Valid, true, "should have been true")
	assert.Equal(v.Bool, false, "should have been false")
}

func TestHash(t *testing.T) {
	assert := assert.New(t)
	v := HashStrings("1")
	assert.Equal(v, "6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b", "should have been 6b86b273ff34fce19d6b804eff5a3f5747ada4eaa22f1d49c01e52ddb7875b4b")
	v = HashStrings("1", "2")
	assert.Equal(v, "6b51d431df5d7f141cbececcf79edf3dd861c3b4069f0b11661a3eefacbba918", "should have been 6b51d431df5d7f141cbececcf79edf3dd861c3b4069f0b11661a3eefacbba918")
}

func TestGeometry(t *testing.T) {
	assert := assert.New(t)
	g := ToGeometry("POINT(-122.3890954 37.6145378)")
	assert.Equal(g.String(), "latitude:37.614536 longitude:-122.3891 ")
	assert.Equal(g.Latitude, float32(37.614536))
	assert.Equal(g.Longitude, float32(-122.3891))
}

func TestTimestamp(t *testing.T) {
	assert := assert.New(t)
	tv := time.Now()
	ts := ToTimestamp(ToSQLDate(tv))
	assert.Equal(ts.Nanos, int32(tv.Nanosecond()))
	assert.Equal(ts.Seconds, tv.Unix())

	dt := ToTimestamp(mysql.NullTime{Time: time.Now(), Valid: true})
	assert.NotNil(dt)
	sdt := ToSQLDate(dt)
	assert.Equal(sdt.Valid, true)
}
