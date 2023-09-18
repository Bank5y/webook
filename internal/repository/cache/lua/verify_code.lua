local key = KEYS[1]
-- 用户输入的code
local expectedCode = ARGV[1]
local code = redis.call("get",key)
local cntKey = key..":cnt"
-- 转换为数组
local cnt = tonumber(redis.call("get", cntKey))
if cnt==nil or cnt <= 0 then
    return -1
elseif expectedCode == code then
    redis.call("set",cntKey,-1)
    return 0
else
    redis.call("decr",cntKey)
    return -2
end
