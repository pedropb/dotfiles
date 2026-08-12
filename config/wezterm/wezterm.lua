local wezterm = require("wezterm")
local act = wezterm.action
local config = {}

config.color_scheme = "Catppuccin Mocha"
config.font = wezterm.font_with_fallback({ "JetBrains Mono", "Symbols Nerd Font Mono" })
config.font_size = 15.0
config.line_height = 1.1
config.window_background_opacity = 0.93
if wezterm.target_triple:find("apple") then
  config.macos_window_background_blur = 20
end
config.window_decorations = "RESIZE"
config.window_padding = { left = 14, right = 14, top = 12, bottom = 12 }
config.inactive_pane_hsb = { saturation = 0.85, brightness = 0.72 }
config.default_cursor_style = "SteadyBar"

config.use_fancy_tab_bar = false
config.enable_tab_bar = true
config.hide_tab_bar_if_only_one_tab = false
config.tab_max_width = 32
config.colors = {
  tab_bar = {
    background = "#11111b",
    active_tab = { bg_color = "#cba6f7", fg_color = "#11111b", intensity = "Bold" },
    inactive_tab = { bg_color = "#1e1e2e", fg_color = "#a6adc8" },
    inactive_tab_hover = { bg_color = "#313244", fg_color = "#cdd6f4" },
    new_tab = { bg_color = "#11111b", fg_color = "#a6adc8" },
    new_tab_hover = { bg_color = "#313244", fg_color = "#cdd6f4" },
  },
}

-- Ctrl-b is a tmux-style prefix. Vim's split and navigation keys apply after it.
config.leader = { key = "b", mods = "CTRL", timeout_milliseconds = 1000 }
config.keys = {
  -- Preserve Ctrl-b for readline/zsh when it is pressed twice.
  { key = "b", mods = "LEADER", action = act.SendKey({ key = "b", mods = "CTRL" }) },

  -- Pane lifecycle: Ctrl-b v makes a vertical split; Ctrl-b s makes a horizontal split.
  { key = "v", mods = "LEADER", action = act.SplitHorizontal({ domain = "CurrentPaneDomain" }) },
  { key = "s", mods = "LEADER", action = act.SplitVertical({ domain = "CurrentPaneDomain" }) },
  { key = "x", mods = "LEADER", action = act.CloseCurrentPane({ confirm = true }) },

  -- Vim-style pane focus and resizing.
  { key = "h", mods = "LEADER", action = act.ActivatePaneDirection("Left") },
  { key = "j", mods = "LEADER", action = act.ActivatePaneDirection("Down") },
  { key = "k", mods = "LEADER", action = act.ActivatePaneDirection("Up") },
  { key = "l", mods = "LEADER", action = act.ActivatePaneDirection("Right") },
  { key = "H", mods = "LEADER|SHIFT", action = act.AdjustPaneSize({ "Left", 5 }) },
  { key = "J", mods = "LEADER|SHIFT", action = act.AdjustPaneSize({ "Down", 5 }) },
  { key = "K", mods = "LEADER|SHIFT", action = act.AdjustPaneSize({ "Up", 5 }) },
  { key = "L", mods = "LEADER|SHIFT", action = act.AdjustPaneSize({ "Right", 5 }) },

  -- Tab lifecycle and navigation.
  { key = "t", mods = "LEADER", action = act.SpawnTab("CurrentPaneDomain") },
  { key = "[", mods = "LEADER", action = act.ActivateTabRelative(-1) },
  { key = "]", mods = "LEADER", action = act.ActivateTabRelative(1) },
  { key = "1", mods = "LEADER", action = act.ActivateTab(0) },
  { key = "2", mods = "LEADER", action = act.ActivateTab(1) },
  { key = "3", mods = "LEADER", action = act.ActivateTab(2) },
  { key = "4", mods = "LEADER", action = act.ActivateTab(3) },
  { key = "5", mods = "LEADER", action = act.ActivateTab(4) },
  { key = "6", mods = "LEADER", action = act.ActivateTab(5) },
  { key = "7", mods = "LEADER", action = act.ActivateTab(6) },
  { key = "8", mods = "LEADER", action = act.ActivateTab(7) },
  { key = "9", mods = "LEADER", action = act.ActivateTab(8) },
}

return config
