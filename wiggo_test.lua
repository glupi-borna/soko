local radius = 0

function SliderWidget()
    if TextButton("-") then radius = radius - 1 end
    Invisible(Px(8))

    radius = Slider(radius, 0, 32)
    local slider = UI().Last
    slider.Size.W = Fr(1)
    slider.Size.H = Px(30)

    Invisible(Px(8))
    if TextButton("+") then radius = radius + 1 end
end

function CloseRow()
    if TextButton("Close") then
        Close()
    end
end

function frame()
    UI().Root.Style = Style{
        Background = RGBA(255, 0, 0, 255),
        CornerRadius = Animate(radius, "radius")
    }

    Text("Corner radius = " .. tostring(math.floor(radius+0.5))).Style = Style{
        Foreground = ColHex(0xffffffff)
    }

    for n in With(Row()) do
        n.Size.W = Fr(1)
        SliderWidget()
    end

    for n in With(Row()) do
        CloseRow()
    end
end
