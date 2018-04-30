e2e: 
	 -killall -q user
	 @echo "Recreating table"
	 psql -U ${DB_USERNAME} ${DB_NAME} < sql/01_user.sql

	 @echo 
	 @echo "Build & run golang app"
	 @time go build -o user && ./user > /dev/null &
	 @sleep 1

	 @echo 
	 @echo "Running a tests"
	 -go test -v main_test.go
	 @killall user
