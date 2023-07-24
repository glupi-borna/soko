local volume = Volume()
local players = nil
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

function PlayerArtist()
    local p = GetPlayer()
    if p == nil then return "" end
    local a, _ = p:GetArtists()
    return a[1]
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

function PlayerArtUrl()
    local p = GetPlayer()
    if p == nil then return "" end
    local s, _ = p:GetArtUrl()
    return s
end

function PlayerArtUrl()
    local p = GetPlayer()
    if p == nil then return "" end
    local s, _ = p:GetArtUrl()
    return s
end

function PlayerVolume()
    local p = GetPlayer()
    if p == nil then return "" end
    local v, _ = p:GetVolume()
    return v
end

function PlayerSetVolume(vol)
    local p = GetPlayer()
    if p == nil then return "" end
    p:SetVolume(vol)
end

function PlayerVolumeSupported()
    local p = GetPlayer()
    if p == nil then return "" end
    local _, err = p:GetVolume()
    return err == nil
end

local up, down, latency = NET.UpDownLatency(false, true)

function frame()
    refreshPlayers()

    local root = UI().Root
    root.Style.Font = "Ubuntu"
    root.Style.FontSize = 16
    root.Style.Foreground.Normal = ColHex(0xf8f8f2ff)
    root.Style.Background.Normal = ColHex(0x1B1C2500)

    local title = PlayerTitle()
    if title ~= "" then
        for n in With(Row()) do
            local img = Image(PlayerArtUrl())
            n.Padding = Padding(0, 0, 0, 8)
            img.Size.W = Em(4)
            img.Size.H = Em(4)
            n.Size.W = ChildrenSize()

            for n in With(Column()) do
                n.Padding.Top = 0
                n.Size.H = ChildrenSize()

                if TextButton(PlayerStatus()) then
                    PlayerPlayPause()
                end
                local button = UI().Last.Parent
                button.Style = root.Style:Copy()
                button.Style.Background.Normal = ColHex(0xf1fa8cff)
                button.Style.Foreground.Normal = ColHex(0x1B1C2500)
                button.Style.Background.Active = ColHex(0xC7D07Aff)
                button.Style.Foreground.Active = ColHex(0x1B1C2500)

                Marquee(PlayerArtist(), 15).Padding.Top=4
                UI().Last.Size.W = Em(10)
                Marquee(title, 15).Padding.Top = 4
                UI().Last.Size.W = Em(10)
            end
        end
    end

    local slider = nil
    if PlayerVolumeSupported() then
        volume = PlayerVolume()
        local old_vol = volume
        volume = Slider(volume, 0, 1)
        slider = UI().Last
        if volume ~= old_vol then PlayerSetVolume(volume) end
    else
        volume = Volume()
        local old_vol = volume
        volume = Slider(volume, 0, 1)
        slider = UI().Last
        if volume ~= old_vol then SetVolume(volume) end
    end

    slider.Style.CornerRadius.Normal = 0
    slider.Style.CornerRadius.Active = 0
    slider.Size.W = Em(16)
    slider.Size.H = Em(1)

    -- Text("Volume")
    -- UI().Last.Padding = Padding(8, 4)

    -- Text(Duration(UI().Delta))

    -- up, down, latency = NET.UpDownLatency(false, Tick(5))

    -- Text(Bytes(up))
    -- Text(Bytes(down))
    -- Text(Duration(latency))
end
