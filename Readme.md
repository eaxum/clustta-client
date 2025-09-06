
# Setting up the development environment

To run the development environment, we need a number of dependencies:
- Go
- Wails3
- Air
- NSIS (for Windows builds)
- Protocol Buffers (Protoc & PBJS)


## Go
Install `Go` for your target OS following instructions from the [Official Documentation](https://go.dev/doc/install)

## Wails3
Install the Wails CLI using Go Modules, run the following command:

```bash
go install -v github.com/wailsapp/wails/v3/cmd/wails3@latest
```

## Air
To run the development server, we will use `Air`. Install using Go Modules:
```bash
go install github.com/air-verse/air@latest
```

## NSIS
install NSIS https://nsis.sourceforge.io/Download and add to path

## Protocol buffers
To transmit data efficiently, Clustta uses Protocol Buffers to serialize data for transmission.

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

```bash
protoc --go_out=. .\internal\repository\schema.proto 
```

```bash
pbjs -t static-module -w es6 --keep-case -o .\frontend\src\lib\repositorypb.js .\internal\repository\schema.proto
```

<br>

# Running the development environment

## Development Client
To initiate a development instance of the client, run:

```bash
make dev
```

>  Optional
>
> Edit the build variables in `build/platform/Taskfile.yml`

This will determine if the development build should connect to the `dev-server` and `dev-studio`

```bash
-ldflags="-X clustta/internal/constants.host=http://127.0.0.1:5000"
```

## Development Server and Studio
If you enabled connecting to the local `studio` and `server` above, you should initiate these as well in this order:

```bash
make dev-server
```

```bash
make dev-studio
```

You will need a `StudioKey`for the `dev-studio` to run successfully. See [Getting a StudioKey](#getting-a-studiokey-for-development).
<br>

# The different components

For any clustta client to work, there are three crucial parts
1. The Global server
2. The Studio server
3. The Client application

We will go over building and or deploying them in the following sections

## The Global Server
This is where all of the data for users and registered studios are stored globally. It acts as an authentication server for users as well as a directory that maps where each server is.

We can deploy this by directly clonning this repository onto the server
```bash
git clone https://github.com/eaxum/clustta.git
```

If it already exists, then update the repo
```bash
git pull
```
Add the .env file ??
```bash
nano ./cmd/server/.env
```

Build/update the application with docker
```bash
docker compose -f ./cmd/server/compose.yml up -d --build
```

Build the studio key into a linux binary so you can easily generate keys when registering new studios:

```bash
go build ./cmd/studio_key/
```

### Getting a StudioKey(Global)
When registering a studio, get the `StudioKey` by running the binary providing two arguments:
- The `clusta.db` path
- The `StudioName` (The studio must have been registered from the dashboard beforehand and must match the exact format - casing and spacing).

```bash
./studio-key /path/to/clustta.db StudioName
```

### Getting a StudioKey for Development
When registering a studio for development, get the `StudioKey` by running the go file at `cmd\studio_key\main.go` providing two arguments:
- The `clusta.db` path
- The `StudioName` (The studio must have been registered from the dashboard beforehand and must match the exact format - casing and spacing).

```bash
go run ./cmd/studio_key ./data/clustta.db StudioName
```

This will print the `StudioKey` to the terminal. Copy and save it for use in [Setting up the variables](#setting-up-the-variables).

### Local studio config
After getting a studio key, you need to update the `server_name` and `studio_api_key` in the `studio_config.json`.

## The Studio Server
This is the private instance of any Clustta studio server. All the studio projects are stored here and the machine's IP is registered on the Clustta Global server.

When a client attempts to reach this server, it first accesses the global server which then routes it to this machine's IP address.

### Building and pushing the Docker image
For privacy reasons, it is deployed using a docker container from the built GO application:

Build the docker image and tag it with the appropriate version
```bash
docker build -f .\cmd\studio_server\Dockerfile -t eaxum/clustta:latest -t eaxum/clustta:x.x.xx .
```

Login to docker hub
```bash
docker login -u eaxum -p password
```

Push the tagged image so it's accessible from any machine on deployment
```bash
docker push eaxum/clustta:latest
```
### Setting up the machine
If docker isnt already installed, install docker and all of it's necessary dependencies on the new machine. Copy and paste the following script into the termial. It will:
- Update system packages
- Install Docker CE (Community Edition)
- Install latest Docker Compose
- Add current user to docker group
- Enable Docker to start on boot
- Clean up unnecessary packages

```bash
sudo apt update && sudo apt upgrade -y && sudo apt install -y apt-transport-https ca-certificates curl software-properties-common && curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo apt-key add - && sudo add-apt-repository "deb [arch=amd64] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" && sudo apt update && sudo apt install -y docker-ce && sudo systemctl enable docker && sudo usermod -aG docker $USER && mkdir -p ~/.docker/cli-plugins/ && curl -SL "https://github.com/docker/compose/releases/latest/download/docker-compose-linux-x86_64" -o ~/.docker/cli-plugins/docker-compose && chmod +x ~/.docker/cli-plugins/docker-compose && sudo apt autoremove -y && echo 
```

After the script completes, you must log out and log back in for Docker group membership to take effect.

### Setting up the variables
Create a directory to place the `docker-compose.yml` and `.env` files

```bash
mkdir clustta-docker
```
Create a `docker-compose.yml`. Copy and paste the contents of `cmd\studio_server\compose.yml` into it.

Create the `.env` file and add the following parameters:

```bash
DATA_FOLDER=/home/server-user-name/data/
PROJECTS_FOLDER=/home/server-user-name/projects/
CLUSTTA_STUDIO_API_KEY=StudioKey
CLUSTTA_SERVER_NAME=StudioName
CLUSTTA_SERVER_URL=http://host-ip/clustta
```
`server-user-name` : the username to the host machine.

`StudioKey` : the key generated earlier in [Getting a StudioKey](#getting-a-studiokey).

`StudioName` : The registered studio name exactly as formatted on creation.

`host-ip` : The IP address of the host machine.

Provide the eaxum docker hub credentials
```bash
docker login
```

Compose the docker file
```bash
docker compose up
```

<br>


> ⚠️ NOTE
>
> Ensure the ports `80` and `443` are enabled on your server host else the client can't reach it.

# The Client application
This is the desktop application from which the users can acces and manage their working files.
The production version is distributed through the Microsoft and Apple stores.

This section will go over building for production and deployment.


# Building the client executable
To build the client for deployment, run:

```bash
make build
```

This will execute the `build` command on the `Makefile` depending on the development OS.

## Windows
For windows, it will: 
1. Output an `exe` file into `.\bin`. You can run this executable on any windows machine even starting the development environment.
2. Invoke the `MsixPackagingTool.exe` to build the `msix` package using the parameters set in the `.\Clustta_template.xml`. 
3. Output the MSIX file into  `.\bin\msix`. This is the version that will be submitted to the Microsoft Store through the [Partner Center](https://partner.microsoft.com).


> ⚠️ NOTE
>
> For `2` and `3` to work, you must have installed [NSIS](https://nsis.sourceforge.io/Download) and added it to PATH.

<br>

> ⚠️ NOTE
>
> The `Installer Path` parameter in the `.\Clustta_template.xml` needs to be hardcoded as for some reason, it doesn't recognize relative paths.

<br>

```bash
<Installer Path="C:\path\to\clustta\bin\Clustta-amd64-installer.exe" Arguments="/S" />
```

## Updating the versions
To bump an update, edit the `version` in these files:

`Clustta_template.xml`

`build/windows/info.json`

`build/windows/nsis/wails_tools.nsh`

`commands.txt`

`frontend/src/services/utils.js`

The current format is `x.x.xx`








