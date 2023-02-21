# mcbot

This bot uses the Discord interactions API, AWS Lambdas, and AWS API Gateway.

## Who's Online

This feature shows users in a Discord channel who is currently online in a
Minecraft server.

In order to send/update the status message, the bot uses Discord's
[Message Commands](https://discord.com/developers/docs/interactions/application-commands#message-commands)
feature. The user right clicks on the message and uses the context menu to
interact with the feature.

### Initialization

Initialization expects the user to first send a message to the desired channel
with information about the server. The format looks like:

```json
{
  "serverName": "froggyland",
  "serverHost": "mc.froggyfren.com"
}
```

It's recommended that this message is sent to a channel created specifically
for the who's online feature, and locked to only allow the bot to post to the
channel.

Once the message is sent, right click the message and use the context menu to
trigger the who's online feature. The bot will read the configuration message,
query the server, and update the message with who's online!

### Updates

Right click the message and use the context menu to update the who's online
message.
