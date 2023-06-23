local count = 0
function frame()
    if Button("Click me") then
        count = count + 1
        print("Clicked " .. count .. " times")
    end
end
