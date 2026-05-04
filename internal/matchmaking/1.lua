local function matchUsers(queueKey,pubSubChannel,minUsers, minScore, lobbyID, userScore, userID)
    local users = redis.call('ZRANGEBYSCORE', queueKey, minScore, '+inf', 'LIMIT', 0, minUsers)
    if #users >= minUsers then
        table.insert(users, userID)
        redis.call('ZREM', queueKey, unpack(users))
        local lobby = {
            id = lobbyID,
            participants = users,
            created_at = userScore
        }
        local lobbyJson = cjson.encode(lobby)
        redis.call('JSON.SET', 'lobby:' .. lobbyID, '.', lobbyJson)
        redis.call('PUBLISH', pubSubChannel, lobbyID .. ':' .. table.concat(users, ','))
        return {true, lobbyID, users}
    else
        redis.call('ZADD', queueKey, userScore, userID)
        return {false}
    end
end


local queueKey = KEYS[1]
local pubSubChannel = KEYS[2]

local minUsers = tonumber(ARGV[1])
local minScore = tonumber(ARGV[2])
local lobbyID = tonumber(ARGV[3])
local userID = tonumber(ARGV[4])
local userScore = tonumber(ARGV[5])
return matchUsers(queueKey, pubSubChannel, minUsers, minScore, lobbyID, userScore, userID)