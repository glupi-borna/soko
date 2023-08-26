local volume = Volume()
local players = nil
local player = nil
function refreshPlayers()
    if players == nil or Tick(5) then
        local err = nil
        player = nil
        players, err = Players()
        if err ~= nil then
            print(err:Error())
        else
            if #players > 0 then player = players[1] end
        end
    end
end

local up, down, latency = NET.UpDownLatency(false, true)

function frame()
    refreshPlayers()

    local root = UI().Root
    root.Style.Font = "Ubuntu"
    root.Style.FontSize = 16
    root.Style.Foreground.Normal = ColHex(0xf8f8f2ff)
    root.Style.Background.Normal = ColHex(0x1B1C2500)

    local player_volume = nil
    if player ~= nil then
        local track = player:GetTrackInfo()

        player_volume, err = player:GetVolume()
        if err ~= nil then player_volume = nil end

        for n in With(Row()) do
            n.Padding = Padding(0, 0, 0, 8)
            n.Size.W = ChildrenSize()
            local img = nil
            local img_size = nil

            if track.ArtUrl ~= "" then
                img = Image(track.ArtUrl)
                img_size = Animate(4, "img_size")
            else
                img = Invisible(Em(0))
                img_size = Animate(0, "img_size")
            end

            img.Size.W = Em(img_size)
            img.Size.H = Em(img_size)

            for n in With(Column()) do
                n.Padding.Top = 0

                if img_size < 0.5 then
                    n.Padding.Left = 0
                end

                n.Size.H = ChildrenSize()

                if TextButton(player:GetPlaybackStatus() or "?") then
                    player:PlayPause()
                end

                local button = UI().Last.Parent
                button.Style = root.Style:Copy()
                button.Style.Background.Normal = ColHex(0xf1fa8cff)
                button.Style.Foreground.Normal = ColHex(0x1B1C2500)
                button.Style.Background.Active = ColHex(0xC7D07Aff)
                button.Style.Foreground.Active = ColHex(0x1B1C2500)

                local artists = track.Artists
                local artist = "?"
                if track.Artists ~= nil and #track.Artists > 0 then
                    artist = track.Artists[1] or "?"
                end

                Marquee(artist, 15).Padding.Top=4
                UI().Last.Size.W = Em(14)
                Marquee(track.Title or "", 15).Padding.Top = 4
                UI().Last.Size.W = Em(14)
            end
        end
    end

    local set_volume = nil
    if player_volume ~= nil then
        volume = player_volume
        set_volume = function(vol) player:SetVolume(vol) end
    else
        volume = Volume()
        set_volume = SetVolume
    end

    local old_vol = volume
    volume = Slider(volume, 0, 1)
    if volume ~= old_vol then set_volume(volume) end

    if player ~= nil then
        UI().Last.Size.W = LargestSibling()
    else
        UI().Last.Size.W = Em(13)
    end
end
