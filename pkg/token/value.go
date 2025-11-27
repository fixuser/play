package token

import (
	"encoding"
	"encoding/json"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

var _ encoding.BinaryMarshaler = (*Value)(nil)
var _ encoding.BinaryUnmarshaler = (*Value)(nil)

type Value struct {
	UserId           int64
	AccessToken      string
	RefreshToken     string
	OsType           string
	CreatedAt        time.Time
	TokenExpiredAt   time.Time
	RefreshExpiredAt time.Time
	Extras           []byte
}

// genToken generates a new token string
// 使用uuid但是去掉分隔符号
func genToken() string {
	return strings.ReplaceAll(uuid.New().String(), "-", "")
}

func NewValue(userId int64) *Value {
	return &Value{
		UserId:       userId,
		AccessToken:  genToken(),
		RefreshToken: genToken(),
		CreatedAt:    time.Now(),
	}
}

// Copy creates a copy of the Value
func (v *Value) Refresh() (newVal *Value) {
	if v == nil {
		return nil
	}
	newVal = NewValue(v.UserId)
	newVal.OsType = v.OsType
	newVal.Extras = v.Extras
	newVal.Set("refreshed_at", time.Now().Unix())
	newVal.Set("old_access_token", v.AccessToken)
	newVal.Set("old_refresh_token", v.RefreshToken)
	return
}

// updateExpire sets the token and refresh token expiration based on options
func (v *Value) updateExpire(o *options) {
	now := time.Now()
	if o.tokenExpires > 0 {
		v.TokenExpiredAt = now.Add(o.tokenExpires)
	}
	if o.refreshExpires > 0 {
		v.RefreshExpiredAt = now.Add(o.refreshExpires)
	}
}

// IsTokenExpired checks if the token is expired
func (v *Value) IsTokenExpired() bool {
	if v == nil {
		return true
	}
	return v.TokenExpiredAt.Before(time.Now())
}

// IsRefreshExpired checks if the refresh token is expired
func (v *Value) IsRefreshExpired() bool {
	if v == nil {
		return true
	}
	return v.RefreshExpiredAt.Before(time.Now())
}

// IsTokenValid checks if the token is valid
func (v *Value) IsTokenValid(osType string) bool {
	if v == nil {
		return false
	}
	return !v.IsTokenExpired() && v.AccessToken != "" && v.UserId > 0 && (osType == "" || strings.EqualFold(v.OsType, osType))
}

// Expire immediately expires the token and refresh token
func (v *Value) Expire() {
	now := time.Now()
	if v.TokenExpiredAt.After(now) {
		v.TokenExpiredAt = now
	}
	if v.RefreshExpiredAt.After(now) {
		v.RefreshExpiredAt = now
	}
}

// Get gets the value of the key
func (v *Value) Get(key string) (value gjson.Result) {
	return gjson.GetBytes(v.Extras, key)
}

// Set sets the value of the key
func (v *Value) Set(key string, value any) (err error) {
	v.Extras, err = sjson.SetBytes(v.Extras, key, value)
	return
}

// MarshalBinary implements the encoding.BinaryMarshaler interface
func (v *Value) MarshalBinary() (data []byte, err error) {
	return json.Marshal(v)
}

// UnmarshalBinary implements the encoding.BinaryUnmarshaler interface
func (v *Value) UnmarshalBinary(data []byte) (err error) {
	return json.Unmarshal(data, v)
}
