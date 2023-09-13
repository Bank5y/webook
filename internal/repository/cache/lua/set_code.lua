-- 你的验证码在Redis上的key
-- phone_code:login:xx
local key = KEYS[1]
-- 验证次数 我们一个验证码 最多重复三次 用于记录重复了几次
-- phone_code:login:xx:cnt
local cntKey = key..":cnt"
-- 你的验证码 123456
local val= ARGV[1]
-- 过期时间
local ttl = tonumber(redis.call("ttl",key))
if ttl == -1 then
    -- key存在但是没有过期时间
    -- 手动设置key但是没有过期时间
    return -2
elseif ttl == -2 or ttl<540 then
    redis.call("set",key,val)
    redis.call("expire",key,600)
    redis.call("set",cntKey,3)
    redis.call("expire",cntKey,600)
    return 0
else
    -- 发送过于频繁
    return -1
end