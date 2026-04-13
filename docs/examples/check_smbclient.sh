#!/bin/bash

smbclient=smbclient
username="$1"
password="$2"
host="$3"
share="$4"

output=$(echo 'du' | $smbclient -m SMB2 -U "$username%$password" //$host/$share 2>/dev/null)
exit_code=$?

if [ $exit_code -ne 0 ]; then
    echo "command failed with code $exit_code"
    exit 1
fi

output=$(echo "$output" | grep 'blocks available')
total_blocks=$(echo "$output" | awk '{ print $1 }')
block_size=$(echo "$output" | awk '{ gsub(/\./, "", $5); print $5; }')
free_blocks=$(echo "$output" | awk '{ print $6 }')

total_bytes=$((total_blocks * block_size))
free_bytes=$((free_blocks * block_size))
used_bytes=$((total_bytes - free_bytes))

printf '# HELP smb_share_total_bytes Total capacity of the SMB share (bytes)\n'
printf '# TYPE smb_share_total_bytes gauge\n'
printf 'smb_share_total_bytes{host="%s",share="%s"} %d\n' "$host" "$share" "$total_bytes"

printf '# HELP smb_share_free_bytes Available space on the SMB share (bytes)\n'
printf '# TYPE smb_share_free_bytes gauge\n'
printf 'smb_share_free_bytes{host="%s",share="%s"} %d\n' "$host" "$share" "$free_bytes"

printf '# HELP smb_share_used_bytes Used space on the SMB share (bytes)\n'
printf '# TYPE smb_share_used_bytes gauge\n'
printf 'smb_share_used_bytes{host="%s",share="%s"} %d\n' "$host" "$share" "$used_bytes"

printf '\n'

exit 0