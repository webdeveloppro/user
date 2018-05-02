#/bin/sh

GOOS=freebsd GOARCH=amd64 go build -o user 
# ssh -p8156 admin@144.76.14.199 "~/kill_bitcoin2sql.sh && sleep 2 && rm ~/bitcoin2sql"
sleep 1
cat user | ssh -p8156 admin@144.76.14.199 "cat >> ~/user1 && chmod +x ~/user1"
