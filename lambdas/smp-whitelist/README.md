# smp-whitelist

## discord interaction

```json
{
    "id": "789721296606855188",
    "application_id": "768281647362605066",
    "name": "smp",
    "description": "manage smp",
    "options": [
        {
            "type": 2,
            "name": "whitelist",
            "description": "whitelist management",
            "options": [
                {
                    "type": 1,
                    "name": "add",
                    "description": "add user to whitelist",
                    "options": [
                        {
                            "type": 6,
                            "name": "discord_user",
                            "description": "discord user"
                        }
                    ]
                },
                {
                    "type": 1,
                    "name": "remove",
                    "description": "remove user from whitelist",
                    "options": [
                        {
                            "type": 6,
                            "name": "discord_user",
                            "description": "discord user"
                        }
                    ]
                },
                {
                    "type": 1,
                    "name": "list",
                    "description": "list users on whitelist"
                }
            ]
        }
    ]
}
```
