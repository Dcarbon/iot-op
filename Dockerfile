FROM dcarbon/go-shared as builder

WORKDIR /dcarbon/iot-op
COPY . .

RUN go mod tidy && go build -buildvcs=false -o iot-op && \
    cp  iot-op /usr/bin && \
    echo "Build image successs...!"


FROM dcarbon/dimg:minimal

COPY --from=builder /usr/bin/iot-op /usr/bin/iot-op
ENV GIN_MODE=release
ENV IOT_IMAGE_PATH=/data/iot/image

CMD [ "iot-op" ]