# blurrs

![blur](../../images/blur-processor.png)

wasmVision processor written in Rust that performs a blur.

## How to build

```shell
cargo build --target wasm32-unknown-unknown --release
cp ./target/wasm32-unknown-unknown/release/blurrs.wasm ../
```
