local volume = Volume()
local players = nil
local frameIdx = 0

function refreshPlayers()
    if players == nil or Tick(5) then
        local err = nil
        players, err = Players()
        if err ~= nil then print(err:Error()) end
    end
end

function GetPlayer()
    if players == nil then return nil end
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

function frame()
    refreshPlayers()

    local root = UI().Root
    root.Style.Font = "Ubuntu"
    root.Style.FontSize = 16

    local title = PlayerTitle()
    if title ~= "" then
        for n in With(Row()) do
            n.Size.W = ChildrenSize()
            if TextButton(PlayerStatus()) then
                PlayerPlayPause()
            end
            Marquee(PlayerTitle(), 15)
            UI().Last.Padding = Padding(8, 4)
        end
    end

    Text("Volume")
    UI().Last.Padding = Padding(8, 4)

    local old_vol = volume
    volume = Slider(volume, 0, 1)
    local slider = UI().Last
    slider.Size.W = Em(8)
    slider.Size.H = Em(1)
    if volume ~= old_vol then SetVolume(volume) end

    frameIdx = frameIdx + 1
end
