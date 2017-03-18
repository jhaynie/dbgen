package orm

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/golang/protobuf/ptypes"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
)

func ToString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	if i, ok := v.(int); ok {
		return fmt.Sprintf("%d", i)
	}
	if i, ok := v.(int32); ok {
		return fmt.Sprintf("%d", i)
	}
	if i, ok := v.(int64); ok {
		return fmt.Sprintf("%d", i)
	}
	if f, ok := v.(float32); ok {
		return fmt.Sprintf("%f", f)
	}
	if f, ok := v.(float64); ok {
		return fmt.Sprintf("%f", f)
	}
	if b, ok := v.(bool); ok {
		return fmt.Sprintf("%v", b)
	}
	if t, ok := v.(*time.Time); ok {
		return fmt.Sprintf("%v", t)
	}
	if t, ok := v.(time.Time); ok {
		return fmt.Sprintf("%v", t)
	}
	if d, ok := v.(time.Duration); ok {
		return fmt.Sprintf("%v", d)
	}
	if s, ok := v.(sql.NullString); ok {
		return fmt.Sprintf("%s", s.String)
	}
	if s, ok := v.(*sql.NullString); ok {
		return fmt.Sprintf("%s", s.String)
	}
	if i, ok := v.(sql.NullInt64); ok {
		return fmt.Sprintf("%d", i.Int64)
	}
	if i, ok := v.(*sql.NullInt64); ok {
		return fmt.Sprintf("%d", i.Int64)
	}
	if f, ok := v.(sql.NullFloat64); ok {
		return fmt.Sprintf("%f", f.Float64)
	}
	if f, ok := v.(*sql.NullFloat64); ok {
		return fmt.Sprintf("%f", f.Float64)
	}
	if b, ok := v.(sql.NullBool); ok {
		return fmt.Sprintf("%v", b.Bool)
	}
	if b, ok := v.(*sql.NullBool); ok {
		return fmt.Sprintf("%v", b.Bool)
	}
	if t, ok := v.(mysql.NullTime); ok {
		return fmt.Sprintf("%v", t.Time)
	}
	if t, ok := v.(*mysql.NullTime); ok {
		return fmt.Sprintf("%v", t.Time)
	}
	return fmt.Sprintf("%v", v)
}

func ToSQLString(v string) sql.NullString {
	if v == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: v, Valid: true}
}

func ToSQLDate(v interface{}) mysql.NullTime {
	if v == nil {
		return mysql.NullTime{}
	}
	switch v.(type) {
	case mysql.NullTime:
		return v.(mysql.NullTime)
	case time.Time:
		{
			return mysql.NullTime{v.(time.Time), true}
		}
	case *time.Time:
		{
			return mysql.NullTime{*v.(*time.Time), true}
		}
	case *tspb.Timestamp:
		{
			ts, err := ptypes.Timestamp(v.(*tspb.Timestamp))
			if err != nil {
				return mysql.NullTime{}
			}
			return mysql.NullTime{ts, true}
		}
	case string:
		if v.(string) == "now" {
			return mysql.NullTime{Time: time.Now().UTC(), Valid: true}
		}
		if strings.HasSuffix(v.(string), "Z") {
			v = v.(string)[0 : len(v.(string))-1]
		}
		if strings.Contains(v.(string), "T") {
			date, err := time.Parse("2006-01-02T15:04:05", v.(string))
			if err != nil {
				return mysql.NullTime{}
			}
			return mysql.NullTime{Time: date.UTC(), Valid: true}
		} else {
			if v.(string) == "" {
				return mysql.NullTime{}
			} else {
				date, err := time.Parse("2006-01-02 15:04:05", v.(string))
				if err != nil {
					return mysql.NullTime{}
				}
				return mysql.NullTime{Time: date.UTC(), Valid: true}
			}
		}
	default:
		return mysql.NullTime{}
	}
}

func toInt64(v string) int64 {
	if v, err := strconv.ParseInt(v, 10, 64); err == nil {
		return v
	}
	return 0
}

func toInt32(v string) int32 {
	if v, err := strconv.ParseInt(v, 10, 32); err == nil {
		return int32(v)
	}
	return 0
}

func toFloat64(v string) float64 {
	if v, err := strconv.ParseFloat(v, 64); err == nil {
		return v
	}
	return 0
}

func ToSQLInt64(v interface{}) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	switch v.(type) {
	case sql.NullInt64:
		return v.(sql.NullInt64)
	case string:
		if v.(string) == "" {
			return sql.NullInt64{}
		}
		return sql.NullInt64{Int64: toInt64(v.(string)), Valid: true}
	case json.Number:
		i, err := v.(json.Number).Int64()
		if err != nil {
			i = 0
		}
		return sql.NullInt64{Int64: i, Valid: true}
	case int64:
		{
			return sql.NullInt64{Int64: v.(int64), Valid: true}
		}
	default:
		return sql.NullInt64{Int64: toInt64(fmt.Sprintf("%v", v)), Valid: true}
	}
}

func ToSQLFloat64(v interface{}) sql.NullFloat64 {
	if v == nil {
		return sql.NullFloat64{}
	}
	switch v.(type) {
	case sql.NullFloat64:
		return v.(sql.NullFloat64)
	case string:
		if v.(string) == "" {
			return sql.NullFloat64{}
		}
		return sql.NullFloat64{Float64: toFloat64(v.(string)), Valid: true}
	case json.Number:
		i, err := v.(json.Number).Float64()
		if err != nil {
			i = 0
		}
		return sql.NullFloat64{Float64: i, Valid: true}
	case int64:
		{
			return sql.NullFloat64{Float64: v.(float64), Valid: true}
		}
	default:
		return sql.NullFloat64{Float64: toFloat64(fmt.Sprintf("%v", v)), Valid: true}
	}
}

func ToSQLBool(v interface{}) sql.NullBool {
	if v == nil {
		return sql.NullBool{}
	}
	switch v.(type) {
	case bool:
		{
			return sql.NullBool{v.(bool), true}
		}
	case string:
		s := v.(string)
		if s == "" {
			return sql.NullBool{}
		}
		if s == "true" || s == "1" {
			return sql.NullBool{true, true}
		}
		return sql.NullBool{false, true}
	case json.Number:
		i, err := v.(json.Number).Int64()
		if err != nil {
			i = 0
		}
		return sql.NullBool{i > 0, true}
	case int64:
		{
			return sql.NullBool{v.(int64) > 0, true}
		}
	default:
		return ToSQLBool(fmt.Sprintf("%v", v))
	}
	return sql.NullBool{}
}

func ToSQLBlob(buf []byte) sql.NullString {
	return sql.NullString{string(buf), true}
}

func HashStrings(objects ...string) string {
	h := sha256.New()
	for _, o := range objects {
		io.WriteString(h, o)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func ToGeometry(point string) *Geometry {
	// POINT(-122.3890954 37.6145378)
	if strings.HasPrefix(point, "POINT(") {
		tok := strings.Split(point[6:len(point)-1], " ")
		if len(tok) == 2 {
			return &Geometry{
				Longitude: float32(toFloat64(tok[0])),
				Latitude:  float32(toFloat64(tok[1])),
			}
		}
	}
	return &Geometry{}
}

func ToTimestamp(t mysql.NullTime) *tspb.Timestamp {
	if t.Valid {
		ts, err := ptypes.TimestampProto(t.Time)
		if err == nil {
			return ts
		}
	}
	return nil
}

func ISODate() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func UUID() string {
	return HashStrings(ISODate(), fmt.Sprintf("%d", RandUID()))
}

func RandUID() int32 {
	return int32(rand.Intn(99999999))
}
