-- Match config/cmux/cmux.json and the tmux prefix: Ctrl+B.
-- Rex consumes the prefix while a sequence is pending; Ctrl+B twice sends it
-- to the focused terminal when a program needs the literal key.
rex.bind("ctrl+b>ctrl+b", "pane.send_key", { key = "ctrl+b" })

rex.bind("ctrl+b>v", "pane.split", { direction = "right" })
rex.bind("ctrl+b>s", "pane.split", { direction = "down" })
rex.bind("ctrl+b>h", "pane.focus", { direction = "left" })
rex.bind("ctrl+b>j", "pane.focus", { direction = "down" })
rex.bind("ctrl+b>k", "pane.focus", { direction = "up" })
rex.bind("ctrl+b>l", "pane.focus", { direction = "right" })
rex.bind("ctrl+b>z", "pane.zoom")
rex.bind("ctrl+b>x", "pane.close")
rex.bind("ctrl+b>c", "client.tab.new")
rex.bind("ctrl+b>1", "client.tab.goto", { index = 1 })
