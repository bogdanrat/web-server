echo 'Building images...'

docker image build --build-arg ACCESS_TOKEN_USR=$ACCESS_TOKEN_USR --build-arg ACCESS_TOKEN_PWD=$ACCESS_TOKEN_PWD -t bogdanrat/webserver_coreservice ./service/core
docker image build --build-arg ACCESS_TOKEN_USR=$ACCESS_TOKEN_USR --build-arg ACCESS_TOKEN_PWD=$ACCESS_TOKEN_PWD -t bogdanrat/webserver_authservice ./service/auth
docker image build --build-arg ACCESS_TOKEN_USR=$ACCESS_TOKEN_USR --build-arg ACCESS_TOKEN_PWD=$ACCESS_TOKEN_PWD -t bogdanrat/webserver_storageservice ./service/storage

echo 'Pushing images...'
echo $DOCKER_PASSWORD | docker login --username $DOCKER_USER --password-stdin

docker push bogdanrat/webserver_coreservice
docker push bogdanrat/webserver_authservice
docker push bogdanrat/webserver_storageservice

echo 'Done.'

