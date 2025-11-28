package token

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
)

var (
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrInvalidAccessToken   = errors.New("invalid access token")
	ErrRefreshTokenExpired  = errors.New("refresh token expired")
)

type Token struct {
	rdb             redis.Cmdable
	opts            *options
	refreshTokenKey string
	tokenUniqueKey  string
	tokenDataKey    string
}

func New(rdb redis.Cmdable, opts ...Option) *Token {
	o := new(options)
	o.apply(opts...).setDefault()

	return &Token{
		rdb:             rdb,
		opts:            o,
		tokenDataKey:    o.prefix + ":token:data",
		tokenUniqueKey:  o.prefix + ":token:unique",
		refreshTokenKey: o.prefix + ":token:refresh",
	}
}

// Set sets the token
func (tk *Token) Set(ctx context.Context, val *Value, opts ...Option) (err error) {
	o := *tk.opts
	o.apply(opts...)
	if val.TokenExpiredAt.IsZero() {
		val.TokenExpiredAt = val.CreatedAt.Add(o.tokenExpires)
	}
	if val.RefreshExpiredAt.IsZero() {
		val.RefreshExpiredAt = val.CreatedAt.Add(o.refreshExpires)
	}

	var oldValue Value
	oldAccessToken := tk.rdb.HGet(ctx, tk.tokenUniqueKey, cast.ToString(val.UserId)).Val()
	if oldAccessToken != val.AccessToken {
		err = tk.rdb.HGet(ctx, tk.tokenDataKey, oldAccessToken).Scan(&oldValue)
		if err != nil {
			if err != redis.Nil {
				return
			}
			err = nil
		}
	}

	_, err = tk.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) (err error) {
		// 如果有旧的 token，则让旧 token 记录失效，让refresh token 失效
		if oldValue.UserId > 0 {
			oldValue.Expire()
			pipe.HSet(ctx, tk.tokenDataKey, oldValue.AccessToken, &oldValue)
			pipe.HDel(ctx, tk.refreshTokenKey, oldValue.RefreshToken, val.AccessToken)
		}
		pipe.HSet(ctx, tk.tokenDataKey, val.AccessToken, val)
		pipe.HSet(ctx, tk.refreshTokenKey, val.RefreshToken, val.AccessToken)
		pipe.HSet(ctx, tk.tokenUniqueKey, val.UserId, val.AccessToken)
		return
	})
	return
}

// Update updates the token
func (tk *Token) Update(ctx context.Context, val *Value, opts ...Option) (err error) {
	o := *tk.opts
	o.apply(opts...)

	expiredAt := val.CreatedAt.Add(o.tokenExpires)
	if expiredAt.Sub(val.RefreshExpiredAt) > time.Minute*2 {
		val.TokenExpiredAt = expiredAt
		err = tk.rdb.HSet(ctx, tk.tokenDataKey, val.AccessToken, val).Err()
	}
	return
}

// Remove removes the token
func (tk *Token) Remove(ctx context.Context, userId int64, token string) (err error) {
	if userId > 0 && token == "" { // 通过 userId 查找 token
		token = tk.rdb.HGet(ctx, tk.tokenUniqueKey, cast.ToString(userId)).Val()
		if token == "" {
			return
		}
	}

	val := new(Value)
	err = tk.rdb.HGet(ctx, tk.tokenDataKey, token).Scan(val)
	if err != nil {
		if err == redis.Nil {
			err = nil
		}
		return
	}

	_, err = tk.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) (err error) {
		pipe.HDel(ctx, tk.tokenDataKey, token)
		pipe.HDel(ctx, tk.refreshTokenKey, val.RefreshToken)
		pipe.HDel(ctx, tk.tokenUniqueKey, cast.ToString(val.UserId))
		return
	})
	return
}

// Get gets the token
func (tk *Token) Get(ctx context.Context, token string) (val *Value, err error) {
	val = new(Value)
	err = tk.rdb.HGet(ctx, tk.tokenDataKey, token).Scan(val)
	if err == redis.Nil {
		err = nil
	}
	return
}

// Refresh refreshes the token
// 把老的数据设置为过期，新的数据重新生成
func (tk *Token) Refresh(ctx context.Context, refreshToken string, opts ...Option) (val *Value, err error) {
	accessToken := tk.rdb.HGet(ctx, tk.refreshTokenKey, refreshToken).Val()
	if accessToken == "" {
		err = ErrRefreshTokenNotFound
		return
	}

	val, err = tk.Get(ctx, accessToken)
	if err != nil || val == nil {
		err = ErrInvalidAccessToken
		return
	}

	if val.IsRefreshExpired() {
		err = ErrRefreshTokenExpired
		return
	}

	o := *tk.opts
	o.apply(opts...)

	newVal := val.Refresh()
	newVal.updateExpire(&o)
	val.Expire()

	_, err = tk.rdb.Pipelined(ctx, func(pipe redis.Pipeliner) (err error) {
		pipe.HSet(ctx, tk.tokenDataKey, newVal.AccessToken, newVal)
		pipe.HSet(ctx, tk.refreshTokenKey, newVal.RefreshToken, newVal.AccessToken)
		pipe.HSet(ctx, tk.tokenUniqueKey, newVal.UserId, newVal.AccessToken)

		pipe.HSet(ctx, tk.tokenDataKey, val.AccessToken, val)
		pipe.HDel(ctx, tk.refreshTokenKey, refreshToken)
		return
	})
	if err != nil {
		return nil, err
	}
	return newVal, nil
}
