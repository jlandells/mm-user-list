# mm-user-list

`mm-user-list` is a utility designed for Mattermost administrators to generate a CSV file listing users who are either members of a specific team or not currently in any teams. By default, bot accounts are excluded from the output, but you can choose to include them if needed.

## Features

- **List Users in a Team**: Generate a CSV of users in a specified team.
- **List Users Not in Any Team**: Generate a CSV of users not currently assigned to any team.
- **Include Bot Accounts**: Optionally include bot accounts in the output.
- **CSV Output**: Specify the output file name for the generated CSV.

## Prerequisites

- Mattermost instance with sysadmin access.
- API token with the necessary permissions.

## Download

Pre-built binaries are available for various platforms. Download the appropriate version for your system from the [Releases](https://github.com/jlandells/mm-user-list/releases) page.

1. Go to the [Releases](https://github.com/jlandells/mm-user-list/releases) page.
2. Download the binary for your platform (e.g., `mm-user-list-linux-amd64`, `mm-user-list-windows-amd64.exe`, `mm-user-list-darwin-amd64`).
3. Ensure the binary is executable (Linux/macOS: `chmod +x mm-user-list`).

## Usage

```bash
./mm-user-list [options]
```

### Command Line Options and Environment Variables

You can configure the utility using command line options or environment variables. Command line options will take precedence if both are provided.

| **Command Line**  | **Environment** | **Notes**                                                                 |
|-------------------|-----------------|----------------------------------------------------------------------------|
| `-url`            | `MM_URL`        | **Required**. The Mattermost host that will receive the API requests.      |
| `-scheme`         | `MM_SCHEME`     | `http` / `https`.  Default is `http`.                                      | 
| `-port`           | `MM_PORT`       | The port used to reach the Mattermost instance. Defaults to `8065`.         |
| `-token`          | `MM_TOKEN`      | **Required**. The API token used to access Mattermost. The user **must** have sysadmin rights. |
| `-team`           |                 | The team for which the users should be listed.                             |
| `-not-in-team`    |                 | Produces a list of users not currently in any team. (Only `team` or `not-in-team` can be supplied. Providing both will result in an error.) |
| `-include-bots`   |                 | Includes bot accounts in the output.                                       |
| `-file`           |                 | **Required**. The name of the CSV file for output.                        |
| `-debug`          | `MM_DEBUG`      | Executes the application in debug mode, providing additional output.       |
| `-version`        |                 | Prints the current version and exits.                                     |
| `-help`           |                 | Displays usage instructions and exits.                                    |

### Example Usage

List all users in a specific team:

```bash
./mm-user-list -url=https://mattermost.example.com -scheme=https -token=YOUR_API_TOKEN -team=my-team -file=users.csv
```

List users not in any teams:

```bash
./mm-user-list -url=https://mattermost.example.com -scheme=https -token=YOUR_API_TOKEN -not-in-team -file=no-team-users.csv
```

Include bot accounts in the output:

```bash
./mm-user-list -url=https://mattermost.example.com -port=80 -token=YOUR_API_TOKEN -team=my-team -include-bots -file=users-with-bots.csv
```

### Debug Mode

Enable debug mode for additional logging:

```bash
./mm-user-list -debug -url=https://mattermost.example.com -scheme=https -token=YOUR_API_TOKEN -team=my-team -file=users.csv
```

## Contributing

We welcome contributions from the community! Whether it's a bug report, a feature suggestion, or a pull request, your input is valuable to us. Please feel free to contribute in the following ways:
- **Issues and Pull Requests**: For specific questions, issues, or suggestions for improvements, open an issue or a pull request in this repository.
- **Mattermost Community**: Join the discussion in the [Integrations and Apps](https://community.mattermost.com/core/channels/integrations) channel on the Mattermost Community server.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Contact

For questions, feedback, or contributions regarding this project, please use the following methods:
- **Issues and Pull Requests**: For specific questions, issues, or suggestions for improvements, feel free to open an issue or a pull request in this repository.
- **Mattermost Community**: Join us in the Mattermost Community server, where we discuss all things related to extending Mattermost. You can find me in the channel [Integrations and Apps](https://community.mattermost.com/core/channels/integrations).
- **Social Media**: Follow and message me on Twitter, where I'm [@jlandells](https://twitter.com/jlandells).
