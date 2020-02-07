# ThanksBot
The slack bot to add thanks in workspaces

![ThanksBot Background](https://user-images.githubusercontent.com/2386877/73987755-48472c00-48f6-11ea-8bf3-ac6660cfcf08.png)

Sometimes you just want to thank someone for something quickly to let them know you appreciate something they did. ThanksBot enables this.

## Commands
ThanksBot adds two new commands to your workspace:

### /thanks
Send a thanks to someone else in the workspace. The syntax is is `/thanks <your message>`. Everyone mentioned in the thanks will be sent a message from ThanksBot with the reason.

Running `/thanks` without any other text will return the introduction text.

### /mythanks
Show all the thanks that you've received. It by default returns all the thanks you've received in the past week but will accept parameters: `day`, `week`, `month`, `quarter`, `alltime`

## Building
### Tl;Dr
```bash
# Build just thanks_bot images
docker build -t thanks_bot:latest thanks_bot/

# Build the test infra with the postgres backend
docker-compose up
```
I provide a docker-compose file that spins up test infra which is a good way to test it with the deployment settings outlined below. For production the `Dockerfile` inside `thanks_bot` builds an image based on scratch in a three stage build process that contains almost nothing except thanks_bot.

## Deployment
You will need to create an app inside your workplace (I called mine Thanks). You then will have to create two slack commands (for consistency I recommend them being called `/thanks` and `/mythanks`).

![Slack Commands](https://user-images.githubusercontent.com/2386877/73989291-2cde2000-48fa-11ea-98a9-587011bce9c1.png)

Both should have escape enabled (otherwise the backend won't be able to find who is thanking whom):
![Slack Command Settings](https://user-images.githubusercontent.com/2386877/73989292-2f407a00-48fa-11ea-9b2e-8f8b489cecd9.png)

You'll also need to change the domain to where ever you host your thanks_bot server.

Thanks for stopping by!
