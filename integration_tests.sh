#/bin/bash
source .env.sh

echo "Recreating database"
psql -U $DB_USERNAME $DB_NAME < sql/01_user.sql

echo "Build & run golang app"
go build -o user && ./user > /dev/null &

sleep 2

echo "Test user register"

STATUS=$(http POST http://127.0.0.1:8082/register -h <<< '{"email": "test@test.com", "password": "newpass"}' | grep HTTP/  | cut -d ' ' -f 2)
if [ "$STATUS" != 201 ]; then
  echo "\tRegister right data return wrong status", $STATUS
else 
  echo "\tRegister right data pass"
fi

STATUS=$(http POST http://127.0.0.1:8082/register -h <<< '{"email": "test@test.com", "password": "newpass"}' | grep HTTP/  | cut -d ' ' -f 2)
if [ "$STATUS" != 400 ]; then
  echo "\tRegister wrong data return wrong status", $STATUS
else 
  echo "\tRegister wrong data pass"
fi

STATUS=$(http POST http://127.0.0.1:8082/register -h <<< '{"email": "test1231@test_com", "password": "newpass"}' | grep HTTP/  | cut -d ' ' -f 2)
if [ "$STATUS" != 400 ]; then
  echo "\tRegister wrong data return wrong status", $STATUS
else 
  echo "\tRegister wrong data pass"
fi

STATUS=$(http POST http://127.0.0.1:8082/register -h | grep HTTP/  | cut -d ' ' -f 2)
if [ "$STATUS" != 400 ]; then
  echo "\tRegister empty data return wrong status", $STATUS
else 
  echo "\tRegister empty data pass"
fi

STATUS=$(http GET http://127.0.0.1:8082/profile Authorization:eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJlbWFpbCI6Im5ld0BtYWlsLmNvbSIsImZpcnN0X25hbWUiOiIiLCJsYXN0X25hbWUiOiIifQ.wv7XDm_1m4jS1MW5q_3sE8yxGmGw6Oo7fB0NQ9S6T0A -h | grep HTTP/  | cut -d ' ' -f 2)
if [ "$STATUS" != 200 ]; then
  echo "\tProfile return wrong status", $STATUS
else 
  echo "\tProfile test pass"
fi

STATUS=$(http GET http://127.0.0.1:8082/profile -h | grep HTTP/  | cut -d ' ' -f 2)
if [ "$STATUS" != 401 ]; then
  echo "\tProfile return wrong status", $STATUS
else 
  echo "\tProfile test pass"
fi

killall user
