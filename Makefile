add:
	curl -H "Content-Type: application/json" \
	-H "x-api-key: thisisanapikey" \
	http://localhost:3000/containers/add

run_api:
	go run test-mysql.go 

list:
	curl -s -H "Content-Type: application/json" \
	-H "x-api-key: thisisanapikey" \
	http://localhost:3000/containers/list | json_pp

start:
	curl -s -H "Content-Type: application/json" \
	-H "x-api-key: thisisanapikey" \
	http://localhost:3000/containers/start/ng

stop:
	curl -s -H "Content-Type: application/json" \
	-H "x-api-key: thisisanapikey" \
	http://localhost:3000/containers/stop/ng

login_mysql:
	mysql -h localhost -P 3306 --protocol=tcp -u root -p

rebuild:
	docker-compose build

start_docker:
	docker-compose up --force-recreate
	

