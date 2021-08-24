## Installing dependencies

Do these once:

```
rustup target add x86_64-unknown-linux-musl
```

Install the following C++ tools from **Visual Studio Build Tools 2019**:
- MSVC v142 - VS 2019 C++ x64/x86 ...
- Windows 10 SDK (...)
- C++ CMake tools for Windows
- Testing tools core feature ...

## Compiling and running the examples

See [examples](https://github.com/awslabs/aws-lambda-rust-runtime/tree/master/lambda-runtime/examples) in the **aws-lambda-rust-runtime** GitHub repo.

### Compile the examples on Windows

```
cargo build -p lambda_runtime --example handle-s3 --release --target x86_64-unknown-linux-musl

cargo build --release --target x86_64-unknown-linux-musl --examples
```

### Prepare the package

```
cp ./target/x86_64-unknown-linux-musl/release/examples/handle-s3 ./bootstrap
zip lambda.zip bootstrap
rm bootstrap
```

### Upload the package to AWS Lambda

```
aws lambda update-function-code --region us-west-2 --function-name S3RuntimeTest --zip-file fileb://lambda.zip
```
