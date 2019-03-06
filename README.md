# Mattermost Profanity Filter Plugin (Beta) ![CircleCI branch](https://img.shields.io/circleci/project/github/mattermost/mattermost-plugin-profanity-filter/master.svg)

This plugin allows you to censor profanity on your Mattermost server.

**Supported Mattermost Server Versions: 5.2+**

## Installation

1. Go to the [releases page of this Github repository](https://github.com/mattermost/mattermost-plugin-antivirus/releases) and download the latest release for your Mattermost server.
2. Upload this file in the Mattermost System Console under **System Console > Plugins > Management** to install the plugin. To learn more about how to upload a plugin, [see the documentation](https://docs.mattermost.com/administration/plugins.html#plugin-uploads).
3. Activate the plugin at **System Console > Plugins > Management**.

### Usage

Inside the `/server` directory, you will find the Go files that make up the server-side of the plugin. Within there, build the plugin like you would any other Go application.

To configure what words are censored on your system, edit the `/server/badwords.go` file.
