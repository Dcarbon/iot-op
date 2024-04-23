TAG=dcarbon/iot-op:test

docker build -t $TAG .
if [[ "$1" == "push" ]];then
    docker push $TAG
fi
