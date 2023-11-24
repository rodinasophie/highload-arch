build-server:
	CGO_ENABLED=0 go build -gcflags="all=-N -l" -o bin/social-network -mod vendor main.go

build-client:
	CGO_ENABLED=0 go build -gcflags="all=-N -l" -o bin/social-network-client -mod vendor client/main.go

docker-init:
	docker compose up -d db-leader  && sleep 2;
	docker exec  -it ha-db-leader sh -c "chmod 0600 /root/.pgpass; psql -h localhost -U admin_user -d postgres -c \"drop role if exists replicator; create role replicator with login replication password 'pass';\" ";
	docker exec -d ha-db-leader sh -c "cd /etc/ && ./pg_setup.sh" && sleep 5;
	docker exec -it ha-db-leader sh -c "rm -rf /pgslave; mkdir /pgslave; pg_basebackup -h localhost -D /pgslave -U replicator -w -v -P --wal-method=stream";
	docker compose up  -d db-replica-1;
	docker exec -it ha-db-leader sh -c "psql -U admin_user -f /etc/highload-arch/schema.sql social_net";
	# docker exec -it ha-db-leader sh -c "psql -U admin_user -d social_net -c \"COPY users(first_name, second_name, birthdate, city, biography) FROM '/people.csv' DELIMITER U&'\0009' CSV HEADER;\"";
	#docker exec -it ha-db-leader sh -c "psql -U admin_user -d social_net -c \" UPDATE users SET biography = 'Empty' WHERE biography IS NULL;\"";
	#docker exec -it ha-db-leader sh -c "psql -U admin_user -d social_net -c \"COPY posts(author_user_id, text, created_at, updated_at) FROM '/posts.csv' DELIMITER U&'\0009' CSV HEADER;\"";
	#docker exec -it ha-db-leader sh -c "psql -U admin_user -d social_net -c \"UPDATE posts SET author_user_id = (SELECT id from users ORDER BY random()+(select extract(epoch from created_at)) LIMIT 1);\" ";

docker-cache:
	docker compose up -d db-cache 

docker-backend:
	docker compose up --build -d backend && sleep 5;

docker-run:
	docker exec -d highload-arch-backend sh -c "./bin/social-network" && sleep 5;

docker-reset:
	docker compose down

generate-users:
	python3 data-generator/generate_users.py

generate-posts:
	python3 data-generator/generate_posts.py

init-system:
	python3 data-generator/initialize_system.py

select-prefix:
	python3 perf-testing/prefix_selection.py

perf-test:
	locust -f perf-testing/main.py  --host=http://localhost:8083

load-for-write:
	python3 perf-testing/load.py