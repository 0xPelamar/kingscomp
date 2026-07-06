-- KEYS
local queueKey = KEYS[1]

-- ARGV
local minUsers = tonumber(ARGV[1])
local minScore = tonumber(ARGV[2])
local lobbyID = ARGV[3]            -- UUID string
local userID = tonumber(ARGV[4])
local userScore = tonumber(ARGV[5])          -- Unix timestamp string (current time)

-- We need to fetch 4 OTHER users from the queue to make a lobby of 5
local neededUsers = minUsers - 1

-- 1. Fetch matching users currently in the queue
local matchedUsers = redis.call('ZRANGEBYSCORE', queueKey, minScore, '+inf', 'LIMIT', 0, neededUsers)

-- 2. Check if we found enough users to form a match
if #matchedUsers >= neededUsers then
    -- convert strings to int
    for i, v in ipairs(matchedUsers) do
        matchedUsers[i] = tonumber(v)
    end

    -- 3. Remove ONLY the matched users from the queue
    -- We do this BEFORE adding the new user to the table, and protect against empty unpacks
    if #matchedUsers > 0 then
        redis.call('ZREM', queueKey, unpack(matchedUsers))
    end
    -- 4. Add the current user (who triggered the match) to our matched group
    table.insert(matchedUsers, userID)

--     -- 5. Create the Lobby JSON object
--     local lobby = {
--         id = lobbyID,
--         participants = matchedUsers,
--         created_at = tonumber(userScore), -- Store as a number in the JSON document
--         state = 'created',
--         resigned = cjson.empty_array
--     }
--     local lobbyJson = cjson.encode(lobby)
--     -- 6. Save Lobby to RedisJSON (Note: Requires the RedisJSON module installed on your server)
--     redis.call('JSON.SET', 'lobby:' .. lobbyID, '.', lobbyJson)
--
--     for i, v in ipairs(matchedUsers) do
--         if v ~= userID then
--             local listKey = 'matchmaking:' .. v
--             redis.call('RPUSH', listKey, lobbyID)
--             redis.call('EXPIRE', listKey, 120)
--         end
--     end

    -- Return success, the lobby ID, and the player array
    return {true, lobbyID, matchedUsers}
else
    -- Not enough users found. Add the current user to the queue to wait.
    redis.call('ZADD', queueKey, userScore, userID)

    -- Return failure (0 = false in Redis)
    return {0}
end