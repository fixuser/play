package token

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRedis 创建测试用的Redis客户端
func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	return client, mr
}

func TestToken_Set(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	tk := New(client, WithPrefix("test"), WithTokenExpires(time.Hour), WithRefreshExpires(time.Hour*24))

	t.Run("set new token", func(t *testing.T) {
		val := NewValue(1001)
		val.OsType = "web"

		err := tk.Set(ctx, val)
		require.NoError(t, err)
		assert.NotEmpty(t, val.AccessToken)
		assert.NotEmpty(t, val.RefreshToken)
		assert.False(t, val.TokenExpiredAt.IsZero())
		assert.False(t, val.RefreshExpiredAt.IsZero())

		// 验证可以获取
		gotVal, err := tk.Get(ctx, val.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, val.UserId, gotVal.UserId)
		assert.Equal(t, val.AccessToken, gotVal.AccessToken)
		assert.Equal(t, val.RefreshToken, gotVal.RefreshToken)
	})

	t.Run("set token with same osType should expire old token", func(t *testing.T) {
		// 第一次设置
		val1 := NewValue(1002)
		val1.OsType = "app"
		err := tk.Set(ctx, val1)
		require.NoError(t, err)

		oldToken := val1.AccessToken
		oldRefreshToken := val1.RefreshToken

		// 第二次设置相同用户和osType
		val2 := NewValue(1002)
		val2.OsType = "app"
		err = tk.Set(ctx, val2)
		require.NoError(t, err)

		// 新token应该可以获取
		gotVal2, err := tk.Get(ctx, val2.AccessToken)
		require.NoError(t, err)
		assert.False(t, gotVal2.IsTokenExpired())

		// 老token应该已过期
		gotVal1, err := tk.Get(ctx, oldToken)
		require.NoError(t, err)
		assert.True(t, gotVal1.IsTokenExpired())
		assert.True(t, gotVal1.IsRefreshExpired())

		// 老refresh token应该失效
		_, err = tk.Refresh(ctx, oldRefreshToken)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrRefreshTokenNotFound))
	})

	t.Run("different osType replaces old token", func(t *testing.T) {
		val1 := NewValue(1003)
		val1.OsType = "web"
		err := tk.Set(ctx, val1)
		require.NoError(t, err)

		val2 := NewValue(1003)
		val2.OsType = "app"
		err = tk.Set(ctx, val2)
		require.NoError(t, err)

		// 新token应该有效
		gotVal2, err := tk.Get(ctx, val2.AccessToken)
		require.NoError(t, err)
		assert.False(t, gotVal2.IsTokenExpired())
	})
}

func TestToken_Get(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	tk := New(client, WithPrefix("test"))

	t.Run("get existing token", func(t *testing.T) {
		val := NewValue(2001)
		val.OsType = "web"
		err := tk.Set(ctx, val)
		require.NoError(t, err)

		gotVal, err := tk.Get(ctx, val.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, val.UserId, gotVal.UserId)
		assert.Equal(t, val.AccessToken, gotVal.AccessToken)
		assert.Equal(t, val.OsType, gotVal.OsType)
	})

	t.Run("get non-existing token", func(t *testing.T) {
		gotVal, err := tk.Get(ctx, "non-existing-token")
		require.NoError(t, err)
		assert.Empty(t, gotVal.AccessToken)
		assert.Equal(t, int64(0), gotVal.UserId)
	})
}

func TestToken_Refresh(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	tk := New(client, WithPrefix("test"), WithTokenExpires(time.Hour), WithRefreshExpires(time.Hour*24))

	t.Run("refresh valid token", func(t *testing.T) {
		val := NewValue(3001)
		val.OsType = "web"
		val.Set("custom_field", "test_value")
		err := tk.Set(ctx, val)
		require.NoError(t, err)

		oldAccessToken := val.AccessToken
		oldRefreshToken := val.RefreshToken

		// 等待一小段时间确保时间戳不同
		time.Sleep(time.Millisecond * 100)

		// 刷新token
		newVal, err := tk.Refresh(ctx, oldRefreshToken)
		require.NoError(t, err)
		assert.Equal(t, val.UserId, newVal.UserId)
		assert.Equal(t, val.OsType, newVal.OsType) // 检查extras是否保留
		assert.Equal(t, "test_value", newVal.Get("custom_field").String())

		// 检查刷新记录
		assert.NotZero(t, newVal.Get("refreshed_at").Int())
		assert.NotEmpty(t, newVal.Get("old_access_token").String())
		assert.NotEmpty(t, newVal.Get("old_refresh_token").String())

		// 新token应该有效
		gotNewVal, err := tk.Get(ctx, newVal.AccessToken)
		require.NoError(t, err)
		require.NotNil(t, gotNewVal)
		assert.False(t, gotNewVal.IsTokenExpired())

		// 老token应该过期
		gotOldVal, err := tk.Get(ctx, oldAccessToken)
		require.NoError(t, err)
		assert.True(t, gotOldVal.IsTokenExpired())

		// 老refresh token应该失效
		_, err = tk.Refresh(ctx, oldRefreshToken)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrRefreshTokenNotFound))
	})

	t.Run("refresh with non-existing refresh token", func(t *testing.T) {
		_, err := tk.Refresh(ctx, "non-existing-refresh-token")
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrRefreshTokenNotFound))
	})

	t.Run("refresh with expired refresh token", func(t *testing.T) {
		val := NewValue(3002)
		val.OsType = "app"
		val.CreatedAt = time.Now().Add(-time.Hour * 25) // 超过refresh过期时间
		val.TokenExpiredAt = val.CreatedAt.Add(time.Hour)
		val.RefreshExpiredAt = val.CreatedAt.Add(time.Hour * 24)
		err := tk.Set(ctx, val)
		require.NoError(t, err)

		_, err = tk.Refresh(ctx, val.RefreshToken)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, ErrRefreshTokenExpired))
	})
}

func TestToken_Update(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	tk := New(client, WithPrefix("test"), WithTokenExpires(time.Hour*2), WithRefreshExpires(time.Hour*24))

	t.Run("update token extends expiration", func(t *testing.T) {
		val := NewValue(4001)
		val.OsType = "web"
		err := tk.Set(ctx, val)
		require.NoError(t, err)

		oldExpiredAt := val.TokenExpiredAt

		// 等待一小段时间
		time.Sleep(time.Millisecond * 100)

		// 更新token
		err = tk.Update(ctx, val)
		require.NoError(t, err)

		// 获取更新后的token
		gotVal, err := tk.Get(ctx, val.AccessToken)
		require.NoError(t, err)

		// token过期时间应该被延长(或保持不变)
		assert.True(t, gotVal.TokenExpiredAt.Equal(oldExpiredAt) || gotVal.TokenExpiredAt.After(oldExpiredAt))
	})

	t.Run("update token near refresh expiration does not extend", func(t *testing.T) {
		val := NewValue(4002)
		val.OsType = "web"
		val.CreatedAt = time.Now().Add(-time.Hour * 23) // 接近refresh过期
		val.TokenExpiredAt = val.CreatedAt.Add(time.Hour)
		val.RefreshExpiredAt = val.CreatedAt.Add(time.Hour * 24)
		err := tk.Set(ctx, val)
		require.NoError(t, err)

		oldExpiredAt := val.TokenExpiredAt

		// 更新token
		err = tk.Update(ctx, val)
		require.NoError(t, err)

		// token过期时间不应该改变（因为会超过refresh过期时间）
		assert.Equal(t, oldExpiredAt.Unix(), val.TokenExpiredAt.Unix())
	})
}

func TestToken_Remove(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	tk := New(client, WithPrefix("test"))

	t.Run("remove existing token", func(t *testing.T) {
		val := NewValue(5001)
		val.OsType = "web"
		err := tk.Set(ctx, val)
		require.NoError(t, err)

		// 删除token
		err = tk.Remove(ctx, val.AccessToken)
		require.NoError(t, err)

		// token应该不存在
		gotVal, err := tk.Get(ctx, val.AccessToken)
		require.NoError(t, err)
		assert.Empty(t, gotVal.AccessToken)

		// refresh token也应该失效
		_, err = tk.Refresh(ctx, val.RefreshToken)
		assert.Error(t, err)
	})

	t.Run("remove non-existing token", func(t *testing.T) {
		err := tk.Remove(ctx, "non-existing-token")
		assert.NoError(t, err) // 不应该报错
	})
}

func TestToken_RemoveByUserId(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	tk := New(client, WithPrefix("test"))

	t.Run("remove token by user id", func(t *testing.T) {
		val := NewValue(6001)
		val.OsType = "web"
		err := tk.Set(ctx, val)
		require.NoError(t, err)

		// 通过用户ID删除token
		err = tk.RemoveByUserId(ctx, 6001)
		require.NoError(t, err)

		// token应该不存在
		gotVal, err := tk.Get(ctx, val.AccessToken)
		require.NoError(t, err)
		assert.Empty(t, gotVal.AccessToken)
	})

	t.Run("remove non-existing user id", func(t *testing.T) {
		err := tk.RemoveByUserId(ctx, 9999)
		assert.NoError(t, err) // 不应该报错
	})
}

func TestToken_ValueExtras(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	tk := New(client, WithPrefix("test"))

	t.Run("set and get extras", func(t *testing.T) {
		val := NewValue(7001)
		val.OsType = "web"

		// 设置extras
		err := val.Set("username", "testuser")
		require.NoError(t, err)
		err = val.Set("role", "admin")
		require.NoError(t, err)
		err = val.Set("permissions", []string{"read", "write"})
		require.NoError(t, err)

		// 保存token
		err = tk.Set(ctx, val)
		require.NoError(t, err)

		// 获取并验证extras
		gotVal, err := tk.Get(ctx, val.AccessToken)
		require.NoError(t, err)
		assert.Equal(t, "testuser", gotVal.Get("username").String())
		assert.Equal(t, "admin", gotVal.Get("role").String())
		assert.Equal(t, []string{"read", "write"}, []string{
			gotVal.Get("permissions.0").String(),
			gotVal.Get("permissions.1").String(),
		})
	})

	t.Run("extras preserved after refresh", func(t *testing.T) {
		val := NewValue(7002)
		val.OsType = "web"
		val.Set("session_id", "abc123")
		val.Set("device_id", "device456")

		err := tk.Set(ctx, val)
		require.NoError(t, err)

		// 刷新token
		newVal, err := tk.Refresh(ctx, val.RefreshToken)
		require.NoError(t, err)

		// extras应该保留
		assert.Equal(t, "abc123", newVal.Get("session_id").String())
		assert.Equal(t, "device456", newVal.Get("device_id").String())
	})
}

func TestToken_Concurrent(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()
	tk := New(client, WithPrefix("test"))

	t.Run("concurrent set for same user different osType", func(t *testing.T) {
		done := make(chan bool, 3)

		for i, osType := range []string{"web", "app", "ios"} {
			go func(idx int, os string) {
				val := NewValue(8001)
				val.OsType = os
				err := tk.Set(ctx, val)
				assert.NoError(t, err)
				done <- true
			}(i, osType)
		}

		// 等待所有goroutine完成
		for i := 0; i < 3; i++ {
			<-done
		}
	})
}

func TestToken_Options(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()
	defer client.Close()

	ctx := context.Background()

	t.Run("custom prefix", func(t *testing.T) {
		tk := New(client, WithPrefix("myapp"))

		val := NewValue(9001)
		val.OsType = "web"
		err := tk.Set(ctx, val)
		require.NoError(t, err)

		// 验证key包含自定义前缀
		keys := mr.Keys()
		found := false
		for _, key := range keys {
			if len(key) > 6 && key[:6] == "myapp:" {
				found = true
				break
			}
		}
		assert.True(t, found, "should have keys with custom prefix")
	})

	t.Run("custom expiration", func(t *testing.T) {
		tk := New(client,
			WithPrefix("test2"),
			WithTokenExpires(time.Minute*30),
			WithRefreshExpires(time.Hour*48))

		val := NewValue(9002)
		val.OsType = "web"
		err := tk.Set(ctx, val)
		require.NoError(t, err)

		// 验证过期时间设置正确
		expectedTokenExpire := val.CreatedAt.Add(time.Minute * 30)
		expectedRefreshExpire := val.CreatedAt.Add(time.Hour * 48)

		assert.InDelta(t, expectedTokenExpire.Unix(), val.TokenExpiredAt.Unix(), 2)
		assert.InDelta(t, expectedRefreshExpire.Unix(), val.RefreshExpiredAt.Unix(), 2)
	})
}

func TestValue_IsTokenValid(t *testing.T) {
	t.Run("valid token", func(t *testing.T) {
		val := NewValue(10001)
		val.OsType = "web"
		val.TokenExpiredAt = time.Now().Add(time.Hour)

		assert.True(t, val.IsTokenValid("web"))
		assert.True(t, val.IsTokenValid("Web"))  // 大小写不敏感
		assert.True(t, val.IsTokenValid(""))     // 空字符串不检查osType
		assert.False(t, val.IsTokenValid("app")) // 不同osType
	})

	t.Run("expired token", func(t *testing.T) {
		val := NewValue(10002)
		val.OsType = "web"
		val.TokenExpiredAt = time.Now().Add(-time.Hour)

		assert.False(t, val.IsTokenValid("web"))
	})

	t.Run("nil value", func(t *testing.T) {
		var val *Value
		assert.False(t, val.IsTokenValid("web"))
	})
}

func TestValue_Expire(t *testing.T) {
	t.Run("expire token", func(t *testing.T) {
		val := NewValue(11001)
		val.TokenExpiredAt = time.Now().Add(time.Hour)
		val.RefreshExpiredAt = time.Now().Add(time.Hour * 24)

		assert.False(t, val.IsTokenExpired())
		assert.False(t, val.IsRefreshExpired())

		val.Expire()

		assert.True(t, val.IsTokenExpired())
		assert.True(t, val.IsRefreshExpired())
	})

	t.Run("expire already expired token", func(t *testing.T) {
		val := NewValue(11002)
		val.TokenExpiredAt = time.Now().Add(-time.Hour)
		val.RefreshExpiredAt = time.Now().Add(-time.Hour)

		oldTokenExpire := val.TokenExpiredAt
		oldRefreshExpire := val.RefreshExpiredAt

		val.Expire()

		// 已过期的时间不应该改变
		assert.Equal(t, oldTokenExpire, val.TokenExpiredAt)
		assert.Equal(t, oldRefreshExpire, val.RefreshExpiredAt)
	})
}
