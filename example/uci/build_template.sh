make clean && make

sftp root@192.168.247.141 << EOF
put ubus_server.out ubus_server.out
EOF
