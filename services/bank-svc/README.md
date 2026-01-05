### Summary

This service is designed to do all tasks related to the financial institution. 

There are some features it offers,

0. Adding banks to the system: 
    - Offers a Page to view existing Financial Institutions
    - If yours does not exist, one is able to add a new Financial Institution
    - Entries get saved to the Bank table (BankID, BankName, created_at)

1. User Flow to register their bank + accounts:
    - User can input Bank name, account ID and access token
    - AccessToken gets saved to the 'Access' Table (TokenID, Access_token, UserID, Created_at, Updated_at)
    - Entries get saved to the 'Accounts' Table (UserID, BankID, AccountID, AccountType, TokenID)

2. The service exposes endpoints to fetch information about a user.
    - Provider UserID and fetch their bank information
    - Output: BankID, AccountID, Access_token


Future Scope, 
 - User is able to search from a list of financial institutions. 


 API Design, 

 POST /v1/bank/institutions
    request_body: {
        "bankName": "Bank of America"
    }
    
    response: 
        - 200 OK
        - Other relevant errors 

GET /v1/bank/institutions

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


 POST /v1/bank/accounts
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

 GET /v1/bank/users/{userID}/banks
    response: {
        "user_id": 123,
        "banks": [
            {
                "Name": "Bank Of America",
                "access_token": "12345",
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
