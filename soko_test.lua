local volume = Volume()
local players = nil
local player = nil
function refreshPlayers()
    local one, two = Volume()
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

function PlayerControls()
    local icon = ""
    if player:GetPlaybackStatus() == "Playing" then
        icon = "media-playback-pause"
    else
        icon = "media-playback-start"
    end

    for row in With(Row()) do
        row.Size.W = ChildrenSize()
        if IconButton("media-skip-backward-symbolic") then player:Previous() end
        Invisible(Em(.25))
        if IconButton(icon) then player:PlayPause() end
        Invisible(Em(.25))
        if IconButton("media-skip-forward-symbolic") then player:Next() end
    end
end

function SongInfo()
    refreshPlayers()
    if player == nil then return end
    local track = player:GetTrackInfo()
    if track == nil then return end

    local root = UI().Root

    for info_container in With(Row()) do
        info_container.Padding = Padding(0, 0, 0, 8)
        info_container.Size.W = ChildrenSize()

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

        for text_container in With(Column()) do
            text_container.Padding.Top = 0
            text_container.Size.H = ChildrenSize()
            if img_size < 0.5 then
                text_container.Padding.Left = 0
            end

            -- for button in With(Button()) do
            --     button.Style = root.Style:Copy()
            --     button.Style.Background.Normal = ColHex(0xf1fa8cff)
            --     button.Style.Foreground.Normal = ColHex(0x1B1C2500)
            --     button.Style.Background.Active = ColHex(0xC7D07Aff)
            --     button.Style.Foreground.Active = ColHex(0x1B1C2500)
            -- end

            local artists = track.Artists
            local artist = "Unknown artist"
            if track.Artists ~= nil and #track.Artists > 0 then
                artist = track.Artists[1] or "Unknown artist"
            end

            Marquee(artist, 15).Padding.Top=4
            UI().Last.Size.W = Em(14)
            Marquee(track.Title or "", 15).Padding.Top = 4
            UI().Last.Size.W = Em(14)

            PlayerControls()
        end
    end
end

function frame()
    local root = UI().Root
    root.Style.Font = "Ubuntu"
    root.Style.FontSize = 16
    root.Style.Foreground.Normal = ColHex(0xf8f8f2ff)
    root.Style.Background.Normal = ColHex(0x1B1C2500)
    root.Padding = Padding(8)

    for row in With(Row()) do
        row.Size.W = ChildrenSize()
        row.Padding = Padding(0)

        for col in With(Column()) do
            col.Padding = Padding(0)
            col.Size.W = ChildrenSize()
            col.Size.H = ChildrenSize()
            local volume = Volume()
            local old_vol = volume
            volume, slider = VSlider(volume, 0, 1)
            slider.Style = root.Style:Copy()
            slider.Style.Align = AlignCenter
            slider.Size.W = Px(12)
            slider.Size.H = Em(8)
            if volume ~= old_vol then SetVolume(volume) end

            Invisible(Em(.25))

            local icon = "audio-volume-high-symbolic"
            if IsMuted() then icon = "audio-volume-muted-symbolic" end
            IconButton(icon)
        end

        SongInfo()
    end

    local scroller = ScrollBegin()
        Text("First Hello is really long so that we have to scroll horizontally and stuff")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Hello")
        Text("Last Hello")
    ScrollEnd()
end
