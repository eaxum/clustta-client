
# Quick install: Setting up Clustta studio on your machine

This is the Clustta desktop application that runs natively on your machine. You can work with it natively offline and or if you have any studios you are invited to, you can collaborate with those users.

Simply download the software for your desired OS from:
  - [Microsoft Store](https://apps.microsoft.com/detail/9pnrghgp3lgx?hl=en-us&gl=NG&ocid=pdpshare)
  - [Mac App Store](https://apps.apple.com/us/app/clustta/id6748349288)



<br>


# Development: Setting up and running the environment

To run the development environment, we need a number of dependencies:
- Go
- Wails3
- Air
- Protocol Buffers (Protoc & PBJS)


## Go
Install `Go` for your target OS following instructions from the [Official Documentation](https://go.dev/doc/install)

## Wails3
Install the Wails CLI using Go Modules, run the following command:
Clustta currently uses v3.0.0-alpha.9

```bash
go install -v github.com/wailsapp/wails/v3/cmd/wails3@v3.0.0-alpha.9
```

## Air
To run the development server, we will use `Air`. Install using Go Modules:
```bash
go install github.com/air-verse/air@latest
```

## Protocol buffers
To transmit data efficiently, Clustta uses protocol buffers to serialize data for transmission.

We are using two libraries: `protoc` and `pbjs` to generate the data for Go and Javascript/Typescript respectively.
Install them like so:

Protoc via Go Modules:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest 
```

PBJS via NPM
```bash
npm install -g protobufjs-cli 
```

<br>

Whenever you update `internal\repository\schema.proto` or `internal\repository\proto_helpers.go` , generate the files like so:

For the backend(GO)
```bash
protoc --go_out=. .\internal\repository\schema.proto 
```

For the frontend(JS/TS)
```bash
pbjs -t static-module -w es6 --keep-case -o .\frontend\src\lib\repositorypb.js .\internal\repository\schema.proto
```

<br>

# Running the development environment

## Development Client
To initiate a development instance of the client, run:

```bash
make client
```

>  Optional
>
> Edit the build variables in `build/platform/Taskfile.yml`

This will determine if the development build should connect to the `development server` and `development studio`

```bash
-ldflags="-X clustta/internal/constants.host=http://127.0.0.1:5000"
```

## Development Server and Studio
If you enabled connecting to the local `studio` and `server` above, refer to [Setting up development Studio server](https://github.com/eaxum/clustta-studio/blob/main/Installation.md#development-server-and-studio) and [Setting up development Auth server](https://github.com/eaxum/clustta-server/blob/main/Installation.md#development-server-and-studio)


<br>


# Building and Deployment


To build the client for deployment, run:

```bash
make build
```

This will execute the `build` command on the `Makefile` depending on the development OS.

This will output the executable into the respective OS directories.












