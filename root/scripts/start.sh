#!/bin/bash

if [ ! -f /config/.config/tint2/tint2rc ]; then
    mkdir -p /config/.config/tint2
    cp /etc/xdg/tint2/tint2rc /config/.config/tint2/tint2rc
    sed -i \
        -e 's/^panel_items = .*/panel_items = T/' \
        -e 's/^panel_position = .*/panel_position = top center horizontal/' \
        -e 's/^taskbar_name = .*/taskbar_name = 0/' \
        -e 's/^rounded = .*/rounded = 0/' \
        -e 's/^task_tooltip = .*/task_tooltip = 0/' \
        /config/.config/tint2/tint2rc
fi

nohup tint2 > /dev/null 2>&1 &
nohup manager > /dev/null 2>&1 &
nohup /usr/bin/chrome --app=http://127.0.0.1:10101 > /dev/null 2>&1 &
