TARGET_DIR=("ealiyun" "edingtalk" "ees" "eetcd" "egitlab" "egorm"  "ejenkins"  "ek8s" "ekafka" "elogger"  "emns" "emongo" "eoauth2" "eredis" "esession" "etoken" "ewechat" "erabbitmq")
ROOT=$(pwd)
for dir in ${TARGET_DIR[@]}; do
    cd $dir && go build -v ./...
    cd $ROOT
done