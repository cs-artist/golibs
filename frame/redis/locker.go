// redis locker
package redis

import (
	"errors"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

var client = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379",
	Password: "",
	DB:       0,
})

// lock: value为唯一值（如uuid，orderid等）
func Lock(key, value string, expiration time.Duration) error {
	ok, err := client.SetNX(key, value, expiration).Result()
	if err != nil {
		return err
	}
	if !ok {
		return errors.New("key exist")
	}
	return nil
}

// unlock：使用lua脚本保障操作的原子性
func Unlock(key, value string) error {
	var luaScript = redis.NewScript(`
		if redis.call("get", KEYS[1]) == ARGV[1]   // compare value, if equal then del
		then
			return redis.call("del", KEYS[1])
		else
			return -1
		end
	`)
	ret, err := luaScript.Run(client, []string{key}, value).Int()
	if err != nil {
		return err
	}
	if ret < 0 {
		return errors.New("value not equal")
	} else if ret == 0 {
		return errors.New("del fail")
	}
	return nil
}

func UUID() string {
	return uuid.NewString()
}
