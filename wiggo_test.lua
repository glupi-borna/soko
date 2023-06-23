local count = 0
function frame()
    if Button("Click me") then
        count = count + 1
        print("Clicked " .. count .. " times")
    end

    if Button("Click me 2") then
        count = count + 2
        print("Clicked " .. count .. " times")
    end

    if Button("Die") then
        error("Died")
    end
end
