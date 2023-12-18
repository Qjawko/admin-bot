# Telegram Bot Challenge Reminder

This project is a Telegram Bot built using Go (version 1.21). It sends reminders to users who request to join a certain group.

## Features

- Generates a math challenge for users who request to join the group.
- If user answers correctly, they receive a message indicating that their request is accepted.
- If the user’s response is incorrect, they are notified and asked to try again.
- If a user warns a challenge but does not respond, the bot reminds them after 10 minutes, 30 minutes and 1 hour.
- Handles high loads and can manage reminders for millions of users.
- Uses a distributed task queue (RabbitMQ) to schedule and send reminders, thereby enabling it to work in a distributed environment with multiple instances.

## How it Works

The bot uses the Go Telegram Bot API to interact with users.
A Redis database is used to store user challenges and their correct answers.
RabbitMQ is set up for scheduling challenge reminders for the users after specific time delays, making this system capable of handling large numbers of users.
