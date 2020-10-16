package pkg

var (
	ErrAlreadyAcquiredLock = "来晚了，已被其他协程抢了锁 %s"

	ErrEtcdV3LockFailure = "etcdv3：抢锁失败"

	ErrRedisLockFailure = "redis：抢锁失败"
)
