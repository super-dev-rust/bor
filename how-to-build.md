# How to build

Follow the https://wiki.polygon.technology/docs/maintain/validate/run-validator-binaries with some adjustments:

```shell
# Installing Bor
make all
ln -nfs PATH/bor/build/bin/bor $GOPATH/bor
ln -nfs PATH/bor/build/bin/bootnode $GOPATH/bootnode

# Setting Up Node Files
git clone https://github.com/maticnetwork/launch
mkdir -p ~/node
cp -rf launch/mainnet-v1/sentry/sentry/* ~/node
cp launch/mainnet-v1/service.sh ~/node

# Setting up the network directories
cd ~/node/heimdall
./setup.sh
cd ~/node/bor
mkdir -p ~/.bor/data/bor/
./setup.sh
```
