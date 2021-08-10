echo 'Building images...'

GIT_COMMIT_SHA=$(git rev-parse HEAD)

docker image build --build-arg ACCESS_TOKEN_USR=$ACCESS_TOKEN_USR --build-arg ACCESS_TOKEN_PWD=$ACCESS_TOKEN_PWD -t bogdanrat/webserver_coreservice:latest -t bogdanrat/webserver_coreservice:$GIT_COMMIT_SHA ./service/core
docker image build --build-arg ACCESS_TOKEN_USR=$ACCESS_TOKEN_USR --build-arg ACCESS_TOKEN_PWD=$ACCESS_TOKEN_PWD -t bogdanrat/webserver_authservicelatest -t bogdanrat/webserver_authservice:$GIT_COMMIT_SHA ./service/auth
docker image build --build-arg ACCESS_TOKEN_USR=$ACCESS_TOKEN_USR --build-arg ACCESS_TOKEN_PWD=$ACCESS_TOKEN_PWD -t bogdanrat/webserver_storageservice:latest -t bogdanrat/webserver_storageservice:$GIT_COMMIT_SHA ./service/storage
docker image build -t bogdanrat/webserver_web:latest -t bogdanrat/webserver_web:$GIT_COMMIT_SHA ./web

echo 'Pushing images...'
echo $DOCKER_PASSWORD | docker login --username $DOCKER_USER --password-stdin

docker push bogdanrat/webserver_coreservice:latest
docker push bogdanrat/webserver_coreservice:$GIT_COMMIT_SHA

docker push bogdanrat/webserver_authservice:latest
docker push bogdanrat/webserver_authservice:$GIT_COMMIT_SHA

docker push bogdanrat/webserver_storageservice:latest
docker push bogdanrat/webserver_storageservice:$GIT_COMMIT_SHA

docker push bogdanrat/webserver_web:latest
docker push bogdanrat/webserver_web:$GIT_COMMIT_SHA

echo 'Done.'
