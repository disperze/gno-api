# GNO API

A proxy for ABCI queries and txs endpoint used by keplr wallet.

Testnet demo: [auth/accounts/g1r0yhmapucpgkhzlgn3pgnh0pz2ax43elfdrj9n](https://lcd.gno.tools/cosmos/auth/v1beta1/accounts/g1r0yhmapucpgkhzlgn3pgnh0pz2ax43elfdrj9n)

## Run

```
make build

./build/gnoapi
```

Custom port and cors enabled.

```
./build/gnoapi --port 1317 --cors
```
