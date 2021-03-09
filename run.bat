cd demo/server
go build -o ../../srv.exe

cd ../client

set GOOS=js
set GOARCH=wasm
go build -o ../../static/main.wasm

cd ../../
srv.exe