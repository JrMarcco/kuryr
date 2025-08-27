-- 窗口大小（毫秒）
local window_size = tonumber(ARGV[1])
-- 限流阈值
local threshold = tonumber(ARGV[2])
-- 当前时间戳（毫秒）
local now = tonumber(ARGV[3])
-- 请求 id ( 避免时间戳冲突 )
local request_id = ARGV[4]

-- 窗口开始时间戳
local window_start = now - window_size

local key = KEYS[1]
local cleanup_key = KEYS[2]

-- 使用概率性清理，减少每次请求的清理操作
-- 每 100 个请求或每秒钟清理一次
local last_cleanup = redis.call('GET', cleanup_key)
local should_cleanup = false

if not last_cleanup then
    should_cleanup = true
else
    local time_since_cleanup = now - tonumber(last_cleanup)
    -- 如果距离上次清理超过 1 秒，则执行清理
    if time_since_cleanup > 1000 then
        should_cleanup = true
    else
        -- 使用随机数实现概率性清理 ( 1% 的概率 )
        local random_num = redis.call('TIME')[2] % 100
        if random_num == 0 then
            should_cleanup = true
        end
    end
end

if should_cleanup then
    redis.call('ZREMRANGEBYSCORE', key, '-inf', window_start)
    -- 清理标记保留 2 秒
    redis.call('SET', cleanup_key, now, 'PX', 2000)
end

-- 统计窗口内的请求数
local count = redis.call('ZCOUNT', key, window_start, '+inf')

if tonumber(count) >= threshold then
    -- 请求达到阈值，限流
    return ""
end

-- 请求未达到阈值，放行
-- 使用时间戳和请求ID组合作为成员，避免时间戳冲突
local member = now .. ":" .. request_id
redis.call('ZADD', key, now, member)
-- 设置键的过期时间为窗口大小的两倍，确保数据不会过早删除
redis.call('PEXPIRE', key, window_size * 2)
return "ok"
