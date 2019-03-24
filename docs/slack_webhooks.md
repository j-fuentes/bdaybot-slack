## Configure Slack webhooks

- Go to [api.slack.com/apps](https://api.slack.com/apps) and login with your organization.
- Click on "Create New App"
- Enter a name for your slack app (e.g. "bdaybot") and select the workspace where you want to use this integration.
- You will be presented with an interface where you can customize things like the name of the bot, or its avatar.
- Go to the "Incoming Webhooks" section.
- Click on "Add New Webhook to Workspace". Select the channel or person where the messages will go, and hit "Authorize".
- You would do this twice, one for the happy bday messages and another for the even messages like "No one to say happy bday today" or "Error reading spreadsheet".
- You will end up with something like this:
![Slack webhooks](./docs/imgs/slack_webhooks.png)
- Copy those webhook URLs into your config file.
