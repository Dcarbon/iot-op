TAG=dcarbon/iot-op:v97

docker build -t $TAG .
if [[ "$1" == "push" ]];then
    docker push $TAG
fi
