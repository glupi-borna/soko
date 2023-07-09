local volume = Volume()
local players = nil
local frameIdx = 0

function refreshPlayers()
    if frameIdx % 120 == 0 then
        local err = nil
        players, err = Players()
        if err ~= nil then print(err:Error()) end
    end
end

function GetPlayer()
    local len = #players
    for i=1, len do
        local p = players[i]
        local _, err = p:GetPlaybackStatus()
        if err == nil then return p end
    end
    return nil
end

function PlayerTitle()
    local p = GetPlayer()
    if p == nil then return "" end
    local t, _ = p:GetTitle()
    return t
end

function PlayerPlayPause()
    local p = GetPlayer()
    if p == nil then return end
    p:PlayPause()
end

function PlayerStatus()
    local p = GetPlayer()
    if p == nil then return "" end
    local s, _ = p:GetPlaybackStatus()
    return s
end

function Marquee(text, len, duration)
    local textlen = #text
    if textlen <= len then
        Text(text)
        return
    end

    local loopIdx = frameIdx % 1200

    local quarter = math.floor(duration*0.25)
    local half = math.floor(duration*0.5)

    local progress = 0
    if loopIdx < quarter then
        progress = 0
    elseif loopIdx >= 3*quarter then
        progress = 1
    else
        progress = ((loopIdx - quarter) % half) / half
    end

    local max = textlen - len
    local firstIdx = math.floor(max * progress)

    if progress == 0 then
        Text(text:sub(firstIdx, firstIdx+len-3) .. "...")
    elseif progress == 1 then
        Text("..." .. text:sub(firstIdx, firstIdx+len-3))
    else
        Text("..." .. text:sub(firstIdx, firstIdx+len-6) .. "...")
    end
end

function frame()
    refreshPlayers()

    local root = UI().Root
    root.Size.W = Px(500)
    root.Style.Font = "Ubuntu"
    root.Style.FontSize = 20

    local title = PlayerTitle()
    if title ~= "" then
        for _ in With(Row()) do
            if TextButton(PlayerStatus()) then
                PlayerPlayPause()
            end
            Marquee(PlayerTitle(), 30, 600)
            UI().Last.Padding = Padding(8, 4)
        end
    end

    Text("Volume")
    UI().Last.Padding = Padding(8, 4)

    local old_vol = volume
    volume = Slider(volume, 0, 1)
    local slider = UI().Last
    slider.Size.W = Fr(1)
    slider.Size.H = Px(16)
    if volume ~= old_vol then SetVolume(volume) end

    frameIdx = frameIdx + 1
end
