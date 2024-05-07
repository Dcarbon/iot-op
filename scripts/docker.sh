TAG=dcarbon/iot-op:pro.2

docker build -t $TAG .
if [[ "$1" == "push" ]];then
    docker push $TAG
fi
