### Summary

This service is designed to do all tasks related to the financial institution. 

There are some features it offers,

1. Adding banks to the system: 
    - Offers a Page to view existing Financial Institutions
    - If yours does not exist, one is able to add a new Financial Institution
    - Entries get saved to the Bank table (BankID, BankName, created_at)

2. User Flow to register their bank + accounts:
    - User can input Bank name, account ID and access token
    - AccessToken gets saved to the 'Access' Table (TokenID, Access_token, UserID, Created_at, Updated_at)
    - Entries get saved to the 'Accounts' Table (UserID, BankID, AccountID, AccountType, TokenID)

3. The service exposes endpoints to fetch information about a user.
    - Provider UserID and fetch their bank information
    - Output: BankID, AccountID (access tokens are never returned by this endpoint - see Security below)


Future Scope, 
 - User is able to search from a list of financial institutions. 
 - Support country codes beyond `US` in Plaid Link token creation.


### Plaid Link flow

The intended real (frontend-driven) flow for linking a bank is:

1. Frontend calls `POST /v1/bank/plaid/link-token` with the user's ID and gets back a `link_token`.
2. Frontend passes that `link_token` to Plaid Link, which the user uses to log into their bank.
3. On success, Link's `onSuccess` callback gives the frontend a short-lived `public_token` plus the institution name.
4. Frontend calls `POST /v1/bank/plaid/exchange` with `user_id`, `bank_name`, and `public_token`. The server exchanges it for an access_token, fetches the real account list from Plaid, and persists every account under that Item - the access_token itself is never sent back to the frontend at any point.

`POST /v1/bank/accounts` still exists separately for saving a single account when you already have an access_token in hand (e.g. manual/scripted use) - it is not replaced by the flow above.

### Security

- `POST /v1/bank/accounts` and `POST /v1/bank/plaid/exchange` never trust client-supplied access tokens or account data at face value: both call Plaid's `/accounts/get` server-side to confirm the token is live, and store the account type Plaid reports (not what the client sent).
- Access tokens are encrypted with AES-256-GCM before being written to SQLite (`PLAID_TOKEN_ENCRYPTION_KEY` env var - generate with `openssl rand -base64 32`) and are never returned in any API response.
- Required env vars: `PLAID_CLIENT_ID`, `PLAID_SECRET`, `PLAID_ENV` (`sandbox` or `production`, defaults to `sandbox`), `PLAID_TOKEN_ENCRYPTION_KEY`. See `.env.example` at the repo root. The service fails fast at startup if any are missing or malformed.
- For local testing without a frontend, run:

  ```
  PLAID_CLIENT_ID=... PLAID_SECRET=... go run ./scripts/sandbox_link.go
  ```

  which mints a Plaid Sandbox `public_token` (simulating Link's `onSuccess`) and prints a ready-to-use `POST /v1/bank/plaid/exchange` request.


 API Design, 

 `POST /v1/bank/institutions`

    request_body: {
        "bankName": "Bank of America"
    }
    
    response: 
        - 200 OK
        - Other relevant errors 

`GET /v1/bank/institutions`

    response: {
        "banks": [
            {
                "bank_id": "123",
                "bank_name": "bofa"
            },
            {
                "bank_id": "987",
                "bank_name": "chase"
            }
        ]
    }


 `POST /v1/bank/accounts`

    request_body: {
        "userID": 123,
        "bankName": "Bofa",
        "Accounts": [
            {
                "AccountID": "9876374",
                "AccountType": "Savings"
            }
        ]
        "AccessToken": "<plaid_token>"
    }

    response: 
    - 200 OK
    - Other relevant errors 

 `GET /v1/bank/users/{userID}/banks`

    response: {
        "user_id": 123,
        "banks": [
            {
                "Name": "Bank Of America",
                "Accounts": [
                    {
                        "AccountID": "123456",
                        "Name" : "Savings"
                    },
                    {
                        "AccountID": "789",
                        "Name" : "Checkings"
                    }
                ]
            },
            {
                ...
            }
        ]
    }

`POST /v1/bank/plaid/link-token`

    request_body: {
        "user_id": 123
    }

    response: {
        "link_token": "link-sandbox-...",
        "expiration": "2026-07-17T22:00:00Z"
    }

`POST /v1/bank/plaid/exchange`

    request_body: {
        "user_id": 123,
        "bank_name": "First Platypus Bank",
        "public_token": "public-sandbox-..."
    }

    response: {
        "bank_id": 3,
        "accounts": [
            {"account_id": "...", "account_type": "depository"},
            {"account_id": "...", "account_type": "credit"}
        ]
    }
