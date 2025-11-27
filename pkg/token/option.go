package token

import "time"

type options struct {
	prefix         string
	tokenExpires   time.Duration
	refreshExpires time.Duration
}

// apply apply options
func (o *options) apply(opts ...Option) *options {
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// setDefault default configuration
func (o *options) setDefault() {
	if o.prefix == "" {
		o.prefix = "user"
	}
	if o.tokenExpires <= 0 {
		o.tokenExpires = time.Hour * 2 // 默认2小时
	}
	if o.refreshExpires <= 0 {
		o.refreshExpires = time.Hour * 24 * 7 // 默认7天
	}
}

type Option func(*options)

// WithPrefix sets the prefix
func WithPrefix(prefix string) Option {
	return func(o *options) {
		o.prefix = prefix
	}
}

// WithTokenExpires sets the token expiration time
func WithTokenExpires(d time.Duration) Option {
	return func(o *options) {
		o.tokenExpires = d
	}
}

// WithRefreshExpires sets the refresh token expiration time
func WithRefreshExpires(d time.Duration) Option {
	return func(o *options) {
		o.refreshExpires = d
	}
}
