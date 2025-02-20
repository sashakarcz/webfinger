# WebFinger Server

This is a lightweight WebFinger server implemented in Go. It serves WebFinger responses based on a YAML configuration file.

## Features
- ✅ Serve WebFinger queries for `acct:` and `https:` resources
- ✅ YAML-based user configuration
- ✅ Auto-reload YAML configuration on changes
- ✅ Supports GitHub, Mastodon, Tailscale, OpenID, and more
- ✅ Runs in a minimal Alpine-based Docker container

## Getting Started

### Prerequisites
- Go 1.21 or later
- Docker (optional, for containerized deployment)

### Installation
Clone the repository:
```sh
 git clone https://github.com/sashakarcz/webfinger.git
 cd webfinger
```

### Running Locally
1. Initialize the Go module:
   ```sh
   go mod tidy
   ```
2. Start the server:
   ```sh
   go run main.go
   ```
3. Test the server:
   ```sh
   curl "http://localhost:8000/.well-known/webfinger?resource=acct:sasha@starnix.net"
   ```

### Configuration
Create a `config.yaml` file in the root directory with user profiles:
```yaml
sasha@starnix.net:
  name: "Sasha Karcz"
  avatar: "https://cdn.fosstodon.org/accounts/avatars/107/339/799/857/005/540/original/777632f4b7102d4f.png"
  openid: "https://auth.starnix.net/"
  github: "https://github.com/sashakarcz"
  mastodon: "https://fosstodon.org/@astrognome"
  tailscale: "https://auth.starnix.net/application/o/tailscale/"
  profile: ""
```

### Building and Running with Docker

1. **Build the Docker Image**:
   ```sh
   docker build -t webfinger:latest .
   ```

2. **Run the Container**:
   ```sh
   docker run -d -p 8000:8000 --name webfinger webfinger:latest
   ```

3. **Test the WebFinger API**:
   ```sh
   curl "http://localhost:8000/.well-known/webfinger?resource=acct:sasha@starnix.net"
   ```

### Automatic YAML Reloading
To enable automatic YAML reloading, ensure that your `config.yaml` is mounted as a volume in Docker:
```sh
docker run -d -p 8000:8000 -v $(pwd)/config.yaml:/app/config.yaml --name webfinger webfinger:latest
```

### License
MIT License. See `LICENSE` for details.

### Author
[Sasha Karcz](https://github.com/sashakarcz)


