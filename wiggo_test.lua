local count = 0

function frame()
    With(Row(), function(n)
        n.Size.W = Fr(1)
        if TextButton("A") then
            count = count + 1
            print("Clicked " .. count .. " times")
        end

        Invisible(Fr(1))
        UI().Last.Size.H = Px(0)

        if TextButton("B") then
            count = count + 2
            print("Clicked " .. count .. " times")
        end
    end)

    With(Row(), function(n)
        if TextButton("Close") then
            Close()
        end

        Invisible(Px(8))

        if TextButton("Die") then
            error("Died")
        end
    end)
end
