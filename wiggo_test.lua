local volume = Volume()

function SliderWidget()
    volume = Volume()
    local old_val = volume

    if TextButton("-") then volume = volume - 0.05 end

    Invisible(Px(8))

    volume = Slider(volume, 0, 1)
    local slider = UI().Last
    slider.Size.W = Fr(1)
    slider.Size.H = Px(30)

    Invisible(Px(8))

    if TextButton("+") then volume = volume - 0.05 end

    if volume ~= old_vol then SetVolume(volume) end
end

function CloseRow()
    if TextButton("Close") then
        Close()
    end
end

function frame()
    for n in With(Row()) do
        n.Size.H = Fr(1)

        for n in With(Column()) do
            n.Size.W = Fr(1)
            Text("Volume")

            for n in With(Row()) do
                SliderWidget()
            end
        end

        local old_vol = volume
        volume = VSlider(volume, 0, 1)
        local slider = UI().Last
        slider.Size.W = Px(16)
        slider.Size.H = Fr(1)
        if volume ~= old_vol then SetVolume(volume) end
    end
end
