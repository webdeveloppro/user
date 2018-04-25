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

STATUS=$(http POST http://127.0.0.1:8082/register -h | grep HTTP/  | cut -d ' ' -f 2)
if [ "$STATUS" != 400 ]; then
  echo "\tRegister empty data return wrong status", $STATUS
else 
  echo "\tRegister empty data pass"
fi

killall user
