---
--- Generated by EmmyLua(https://github.com/EmmyLua)
--- Created by xishan.
--- DateTime: 2024/4/22 01:18
---
local key = KEYS[1]
local exists = redis.call('EXISTS', key)
if exists == 0 then
    return 0
end

local rating = tonumber(ARGV[1])
local currentValues = redis.call('HMGET', key, 'score', 'rater_cnt')

local currentScore = tonumber(currentValues[1])
local currentCount = tonumber(currentValues[2])

if currentCount == 1 then
    redis.call('HSET', key, 'score', 0, 'rater_cnt', 0)
    return 1
end
local newCount = currentCount - 1
local newScore = ((currentScore * currentCount) - rating) / newCount
redis.call('HSET', key, 'score', newScore, 'rater_cnt', newCount)
return 1