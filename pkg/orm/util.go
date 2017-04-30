package orm

import (
	"crypto/sha256"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
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

// ToString returns a string representation for the value passed
func ToString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	if s, ok := v.(*string); ok {
		if s == nil {
			return ""
		}
		return *s
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

// JoinAsString will join the array passed as a string which is comma separated
func JoinAsString(v []interface{}) string {
	s := make([]string, 0)
	for _, i := range v {
		s = append(s, ToString(i))
	}
	return strings.Join(s, ", ")
}

// ToSQLString returns a sql.NullString for value passed
func ToSQLString(value interface{}) sql.NullString {
	if v, ok := value.(string); ok {
		if v == "" {
			return sql.NullString{}
		}
		return sql.NullString{String: v, Valid: true}
	}
	if v, ok := value.(*string); ok {
		if v == nil {
			return sql.NullString{}
		}
		return ToSQLString(*v)
	}
	return ToSQLString(fmt.Sprintf("%v", value))
}

// ToSQLDate returns a mysql.NullTime for the value passed
func ToSQLDate(v interface{}) mysql.NullTime {
	if v == nil {
		return mysql.NullTime{}
	}
	switch v.(type) {
	case mysql.NullTime:
		return v.(mysql.NullTime)
	case time.Time:
		{
			return mysql.NullTime{Time: v.(time.Time), Valid: true}
		}
	case *time.Time:
		{
			return mysql.NullTime{Time: *v.(*time.Time), Valid: true}
		}
	case *tspb.Timestamp:
		{
			ts, err := ptypes.Timestamp(v.(*tspb.Timestamp))
			if err != nil {
				return mysql.NullTime{}
			}
			return mysql.NullTime{Time: ts, Valid: true}
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
		}
		if v.(string) == "" {
			return mysql.NullTime{}
		}
		date, err := time.Parse("2006-01-02 15:04:05", v.(string))
		if err != nil {
			return mysql.NullTime{}
		}
		return mysql.NullTime{Time: date.UTC(), Valid: true}
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

// ToSQLInt64 returns a sql.NullInt64 from the value
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
	case int:
		{
			i := v.(int)
			if IsNullInt(int32(i)) {
				return sql.NullInt64{}
			}
			return sql.NullInt64{Int64: int64(i), Valid: true}
		}
	case int32:
		{
			i := v.(int32)
			if IsNullInt(i) {
				return sql.NullInt64{}
			}
			return sql.NullInt64{Int64: int64(i), Valid: true}
		}
	case int64:
		{
			i := v.(int64)
			if IsNullInt(int32(i)) {
				return sql.NullInt64{}
			}
			return sql.NullInt64{Int64: i, Valid: true}
		}
	case *int:
		{
			i := v.(*int)
			if i == nil {
				return sql.NullInt64{}
			}
			return ToSQLInt64(*i)
		}
	case *int32:
		{
			i := v.(*int32)
			if i == nil {
				return sql.NullInt64{}
			}
			return ToSQLInt64(*i)
		}
	case *int64:
		{
			i := v.(*int64)
			if i == nil {
				return sql.NullInt64{}
			}
			return ToSQLInt64(*i)
		}
	default:
		return sql.NullInt64{Int64: toInt64(fmt.Sprintf("%v", v)), Valid: true}
	}
}

// ToSQLFloat64 returns a sql.NullFloat64 from the value
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
	case int:
		{
			i := v.(int)
			return sql.NullFloat64{Float64: float64(i), Valid: true}
		}
	case *int:
		{
			i := v.(*int)
			if i == nil {
				return sql.NullFloat64{}
			}
			return ToSQLFloat64(*i)
		}
	case int32:
		{
			i := v.(int32)
			return sql.NullFloat64{Float64: float64(i), Valid: true}
		}
	case *int32:
		{
			i := v.(*int32)
			if i == nil {
				return sql.NullFloat64{}
			}
			return ToSQLFloat64(*i)
		}
	case int64:
		{
			i := v.(int64)
			return sql.NullFloat64{Float64: float64(i), Valid: true}
		}
	case *int64:
		{
			i := v.(*int64)
			if i == nil {
				return sql.NullFloat64{}
			}
			return ToSQLFloat64(*i)
		}
	case float32:
		{
			f := v.(float32)
			return sql.NullFloat64{Float64: float64(f), Valid: true}
		}
	case float64:
		{
			f := v.(float64)
			return sql.NullFloat64{Float64: f, Valid: true}
		}
	case *float32:
		{
			f := v.(*float32)
			if f == nil {
				return sql.NullFloat64{}
			}
			return ToSQLFloat64(*f)
		}
	case *float64:
		{
			f := v.(*float64)
			if f == nil {
				return sql.NullFloat64{}
			}
			return ToSQLFloat64(*f)
		}
	default:
		return sql.NullFloat64{Float64: toFloat64(fmt.Sprintf("%v", v)), Valid: true}
	}
}

// ToSQLBool returns a sql.NullBool from the value
func ToSQLBool(v interface{}) sql.NullBool {
	if v == nil {
		return sql.NullBool{}
	}
	switch v.(type) {
	case bool:
		{
			return sql.NullBool{Bool: v.(bool), Valid: true}
		}
	case *bool:
		{
			b := v.(*bool)
			if b == nil {
				return sql.NullBool{}
			}
			return ToSQLBool(*b)
		}
	case string:
		s := v.(string)
		if s == "" {
			return sql.NullBool{}
		}
		if s == "true" || s == "1" {
			return sql.NullBool{Bool: true, Valid: true}
		}
		return sql.NullBool{Bool: false, Valid: true}
	case json.Number:
		i, err := v.(json.Number).Int64()
		if err != nil {
			i = 0
		}
		return sql.NullBool{Bool: i > 0, Valid: true}
	case int64:
		{
			return sql.NullBool{Bool: v.(int64) > 0, Valid: true}
		}
	default:
		return ToSQLBool(fmt.Sprintf("%v", v))
	}
}

// ToSQLBlob returns a sql.NullString from the value
func ToSQLBlob(buf []byte) sql.NullString {
	return sql.NullString{String: string(buf), Valid: true}
}

// HashStrings will return a SHA256 hash of the provided arguments
func HashStrings(objects ...string) string {
	h := sha256.New()
	for _, o := range objects {
		io.WriteString(h, o)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

// ToGeometry will return a Geometry from a POINT string
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

// ToTimestampNow returns the proto Timestamp from curent time
func ToTimestampNow() *tspb.Timestamp {
	ts, _ := ptypes.TimestampProto(time.Now())
	return ts
}

// ToTimestamp returns a proto Timestamp from a mysql.NullTime
func ToTimestamp(t mysql.NullTime) *tspb.Timestamp {
	if t.Valid {
		ts, err := ptypes.TimestampProto(t.Time)
		if err == nil {
			return ts
		}
	}
	return nil
}

// ISODate returns a UTC date as string
func ISODate() string {
	return time.Now().UTC().Format(time.RFC3339)
}

// UUID returns a unique ID
func UUID() string {
	return HashStrings(ISODate(), fmt.Sprintf("%d", RandUID()))
}

// RandUID returns a random int32
func RandUID() int32 {
	return int32(rand.Intn(99999999))
}

// NullIntType is a special null type int32
type NullIntType int32

// NullInt32 is a magic int32 which represents null
const NullInt32 NullIntType = -2147483647

// IsNullInt returns true if the value passed is a special null value
func IsNullInt(v int32) bool {
	return NullIntType(v) == NullInt32
}

func (v NullIntType) String() string {
	return fmt.Sprintf("%v", int32(v))
}

// Value will do the proper serialization for SQL inserting
func (v NullIntType) Value() (driver.Value, error) {
	return NullInt32, nil
}

// Scan will do the proper deserialization for SQL inserting
func (v *NullIntType) Scan(value interface{}) error {
	if value == nil {
		*v = NullInt32
		return nil
	}
	if iv, err := driver.Int32.ConvertValue(value); err == nil {
		if value, ok := iv.(int32); ok {
			*v = NullIntType(value)
			return nil
		}
		if value, ok := iv.(int); ok {
			*v = NullIntType(value)
			return nil
		}
		if value, ok := iv.(int64); ok {
			*v = NullIntType(value)
			return nil
		}
	}
	return errors.New("failed to scan NullIntType")
}

// Stringify will return a JSON formatted string. pass an optional second argument to pretty print
func Stringify(v interface{}, opts ...interface{}) string {
	if len(opts) > 0 {
		buf, err := json.MarshalIndent(v, "", "\t")
		if err != nil {
			return ""
		}
		return string(buf)
	}
	buf, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(buf)
}
