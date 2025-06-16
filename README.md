# treex

**treex** is a file viewer that displays annotations visually.

Ever joined a new project and felt lost in a sea of files and directories? `treex` provides a living map of your codebase, helping you and your team understand the architecture at a glance.

![treex screenshot](https://raw.githubusercontent.com/arthur-debert/treex/main/docs/assets/screenshot.png)

Imagine exploring a new project for the first time. Instead of just a list of files, you get this:

```text
my-web-app
├── .github/                CI/CD workflows
│   └── workflows/
│       └── release.yml     Handles automated deployments to production
├── .gitignore
├── Dockerfile              Containerizes the app for production. Uses a multi-stage build.
├── README.md               You are here!
├── api/                    Backend services (Express.js)
│   ├── .info
│   ├── package.json
│   └── server.js           Main API server file. Defines all routes.
├── package.json            Manages Node.js dependencies for both frontend and backend.
└── web/                    Frontend application (React)
    ├── .info
    ├── package.json
    └── src/
        ├── App.js          The root of our React app.
        └── components/
            └── Login.js    The main login component. Connects to the `/api/server.js` endpoint.
```

This annotated view is powered by simple `.info` files you can check into your repository, making project knowledge accessible and easy to maintain.

## How It Works

`treex` looks for `.info` files in the directories it scans. These files contain simple, Markdown-like annotations for files and directories.

Here's the content of the `web/.info` file from the example above:

```
# web/

# This is the main directory for our React single-page application.
# It has its own package.json for managing frontend dependencies.

App.js
The root of our React app.

components/Login.js
The main login component. Connects to the `/api/server.js` endpoint.
```

It's just a path followed by its description. That's it!

## Installation

You can install `treex` using a package manager or by downloading a pre-compiled binary.

### Package Managers

#### Homebrew (macOS / Linux)

If you are on macOS or Linux, you can install `treex` using [Homebrew](https://brew.sh/):

```bash
# First, add the custom tap
brew tap arthur-debert/tools

# Now, install treex
brew install treex
```

#### Deb/APT (Debian / Ubuntu)

If you are on a Debian-based Linux distribution like Ubuntu, you can install `treex` from our APT repository.

*Note: You will need to replace `your-apt-repo.com` with the actual domain of your repository.*

```bash
# 1. Add the repository's GPG key
curl -sS https://your-apt-repo.com/gpg.key | sudo gpg --dearmor -o /usr/share/keyrings/treex-archive-keyring.gpg

# 2. Add the repository to your sources
echo "deb [signed-by=/usr/share/keyrings/treex-archive-keyring.gpg] https://your-apt-repo.com/ ./" | sudo tee /etc/apt/sources.list.d/treex.list > /dev/null

# 3. Update package lists and install treex
sudo apt-get update
sudo apt-get install treex
```

### Manual Installation

You can always download the latest pre-compiled binary for your operating system and architecture from the [GitHub Releases](https://github.com/arthur-debert/treex/releases) page.

1. Download the appropriate archive (e.g., `treex_Linux_x86_64.tar.gz`).
2. Extract the archive: `tar -xzf treex_*.tar.gz`
3. Move the `treex` binary to a directory in your `$PATH`: `sudo mv treex /usr/local/bin/`

### From Source

If you have Go installed, you can build and install `treex` from source:

```bash
go install github.com/arthur-debert/treex/cmd/treex@latest
```

## Usage

```bash
# Show the annotated tree for the current directory
treex

# Show the tree for a specific path
treex path/to/your/project

# Get help on all available flags
treex --help
```

## Development

Interested in contributing? Check out the [Development Guide](docs/DEVELOPMENT.md) to get started.

## Advanced Topics

### Setting Up an APT Repository

To make `treex` installable via APT, you need to host a repository for your `.deb` packages. While a full guide is beyond the scope of this README, here are the general steps:

1. **Generate a GPG Key**: You need a GPG key to sign your repository and packages. This allows `apt` to verify their authenticity.

    ```bash
    gpg --full-generate-key
    ```

    * When prompted, choose **(1) RSA and RSA**.
    * Select a key size of **4096** bits.
    * Set an expiration date (e.g., 1y) or choose no expiration.
    * Fill in your user ID details (name and email).

2. **Export Your Public GPG Key**: You will need to make this key available so users can add it to their system.

    ```bash
    # Replace 'your-email@example.com' with the email you used for the key
    gpg --export --armor your-email@example.com > gpg.key
    ```

3. **Set Up a Repository Server**:
    * Use a tool like [**aptly**](https://www.aptly.info/) or [**reprepro**](https://wiki.debian.org/reprepro) on a server to manage the repository structure.
    * Alternatively, use a service like [**Cloudsmith**](https://cloudsmith.io/) or [**Gemfury**](https://gemfury.com/) to host your packages.

4. **Add Packages to the Repository**: On each new release, you will add the `.deb` file generated by GoReleaser to your repository and update the repository metadata.

5. **Update Your `README.md`**: Once your repository is live, replace the placeholder URLs in the APT installation instructions with your actual repository domain and the path to your `gpg.key`.

## License

[MIT](LICENSE)
