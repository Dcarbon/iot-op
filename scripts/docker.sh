TAG=dcarbon/iot-op

docker build -t $TAG .
if [[ "$1" == "push" ]];then
    docker push $TAG
fi
