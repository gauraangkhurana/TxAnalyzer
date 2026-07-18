### Summary

user-svc manages user identity records - a `user_id` that bank-svc (and later tx-svc) attach data to. This is a personal, single-instance app, so there's no authentication here: no passwords, no sessions. A "user" is just a username tied to an ID.

### API Design

`POST /v1/users`

    request_body: {
        "username": "gauraang"
    }

    response: {
        "user_id": 1,
        "username": "gauraang"
    }

`GET /v1/users/{userID}`

    response: {
        "user_id": 1,
        "username": "gauraang",
        "created_at": "2026-07-17 22:32:20"
    }

    404 if the user doesn't exist.

### Notes

- Shares the same sqlite database as bank-svc (`db/scripts/init.sql`'s `users` table already existed before this service did).
- Required env vars: `DB_PATH`, `PORT` (defaults to `10002`).
