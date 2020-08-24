add:
	curl -H "Content-Type: application/json" \
	-H "x-api-key: thisisanapikey" \
	http://localhost:3000/containers/add

start:
	docker-compose up --force-recreate
	
get:
	curl -s -H "Content-Type: application/json" \
	-H "x-api-key: thisisanapikey" \
	http://localhost:3000/containers/list | json_pp

rebuild:
	docker-compose build

login_mysql:
	mysql -h localhost -P 3306 --protocol=tcp -u root -p
